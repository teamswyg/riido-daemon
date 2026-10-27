package securecredential

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func installationGolden(t *testing.T) (InstallationPublicKey, InstallationProof, []byte) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "native", "cli-auth-helper", "Tests", "testdata", "installation-proof-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		IntentID, Resource, ClientID, RedirectURI, State, CodeChallenge, ChallengeHash, PublicKey, KeyFingerprint, Code, MessageBase64 string
		ExpiresAt                                                                                                                      int64
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	message, err := base64.StdEncoding.DecodeString(fixture.MessageBase64)
	if err != nil {
		t.Fatal(err)
	}
	return InstallationPublicKey{Algorithm: InstallationKeyAlgorithm, PublicKey: fixture.PublicKey, KeyFingerprint: fixture.KeyFingerprint}, InstallationProof{
		Resource: fixture.Resource, ClientID: fixture.ClientID, Intent: fixture.IntentID, ChallengeHash: fixture.ChallengeHash,
		KeyFingerprint: fixture.KeyFingerprint, Code: fixture.Code, RedirectURI: fixture.RedirectURI,
		CodeChallenge: fixture.CodeChallenge, State: fixture.State, ExpiresAt: fixture.ExpiresAt,
	}, message
}

func TestInstallationGoldenMatchesIndependentServerAndSwift(t *testing.T) {
	key, proof, expected := installationGolden(t)
	message, err := proof.message(key, time.Unix(proof.ExpiresAt-300, 0))
	if err != nil || string(message) != string(expected) {
		t.Fatal("server and helper signed bytes differ")
	}
	if installationStorageKey(proof.Resource) != "08eba32e93e89b49cbed7c040f8de7dba724391c4015de3baaa78d9dad03957f" {
		t.Fatal("storage scope differs")
	}
	// This committed signature was produced by the Swift synthetic key fixture;
	// it is public test data, not a signed-helper/Keychain acceptance result.
	raw, err := os.ReadFile(filepath.Join("..", "..", "native", "cli-auth-helper", "Tests", "testdata", "installation-proof-swift.json"))
	if err != nil {
		t.Fatal(err)
	}
	var swift struct {
		Message   string `json:"message"`
		PublicKey string `json:"public_key"`
		Signature string `json:"signature"`
	}
	if json.Unmarshal(raw, &swift) != nil || swift.Message != base64.StdEncoding.EncodeToString(expected) || swift.PublicKey != key.PublicKey {
		t.Fatal("Swift fixture binding differs")
	}
	signature, err := base64.RawURLEncoding.DecodeString(swift.Signature)
	if err != nil {
		t.Fatal(err)
	}
	publicKey, err := key.publicKey()
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(message)
	if !ecdsa.VerifyASN1(publicKey, digest[:], signature) {
		t.Fatal("Swift DER signature does not verify in Go")
	}
}

func installationLiveFixture(t *testing.T) (InstallationPublicKey, InstallationProof, *ecdsa.PrivateKey) {
	t.Helper()
	_, proof, _ := installationGolden(t)
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	publicBytes, err := privateKey.PublicKey.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(publicBytes)
	key := InstallationPublicKey{Algorithm: InstallationKeyAlgorithm, PublicKey: base64.RawURLEncoding.EncodeToString(publicBytes), KeyFingerprint: hex.EncodeToString(digest[:])}
	proof.KeyFingerprint = key.KeyFingerprint
	proof.ExpiresAt = time.Now().Add(5 * time.Minute).Unix()
	return key, proof, privateKey
}

func TestInstallationHelperTypedOperationsAndScope(t *testing.T) {
	key, proof, privateKey := installationLiveFixture(t)
	var calls []request
	s := testStore(func(_ context.Context, _ string, raw []byte) ([]byte, error) {
		var sent request
		if err := json.Unmarshal(raw, &sent); err != nil {
			return nil, err
		}
		calls = append(calls, sent)
		value, err := base64.StdEncoding.DecodeString(sent.Value)
		if err != nil {
			return nil, err
		}
		if sent.Key != installationStorageKey(proof.Resource) || !sent.Interactive {
			t.Fatal("credential key or noninteractive flag used")
		}
		var data any
		switch sent.Operation {
		case "installation-key":
			var payload map[string]string
			if json.Unmarshal(value, &payload) != nil || len(payload) != 2 || payload["resource"] != proof.Resource || payload["client_id"] != "riido-cli" {
				t.Fatal("key request differs")
			}
			data = key
		case "installation-proof":
			var p InstallationProof
			if json.Unmarshal(value, &p) != nil || p != proof {
				t.Fatal("proof request differs")
			}
			message, messageErr := p.message(key, time.Now())
			if messageErr != nil {
				return nil, messageErr
			}
			digest := sha256.Sum256(message)
			signature, signErr := ecdsa.SignASN1(rand.Reader, privateKey, digest[:])
			if signErr != nil {
				return nil, signErr
			}
			data = map[string]string{"algorithm": InstallationKeyAlgorithm, "signature": base64.RawURLEncoding.EncodeToString(signature)}
		default:
			t.Fatal("unexpected generic operation")
		}
		return json.Marshal(map[string]any{"version": 1, "ok": true, "data": data})
	})
	actual, err := s.EnsureInstallationKey(context.Background(), proof.Resource)
	if err != nil || actual != key {
		t.Fatal("key query failed")
	}
	if signature, err := s.SignInstallationProof(context.Background(), key, proof); err != nil || signature == "" {
		t.Fatal("typed proof failed")
	}
	if len(calls) != 2 || s.Key != "account-key-0123456789" {
		t.Fatal("operation changed credential identity")
	}
}

func TestInstallationHelperRefusesWithoutVerificationOrInteraction(t *testing.T) {
	key, proof, _ := installationLiveFixture(t)
	for _, mode := range []string{"noninteractive", "signature", "cancelled"} {
		t.Run(mode, func(t *testing.T) {
			called := false
			s := testStore(func(context.Context, string, []byte) ([]byte, error) { called = true; return nil, nil })
			ctx := context.Background()
			switch mode {
			case "noninteractive":
				s.Interactive = false
			case "signature":
				s.verify = func(context.Context, string) error { return errors.New("private detail") }
			case "cancelled":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			if _, err := s.EnsureInstallationKey(ctx, proof.Resource); err == nil {
				t.Fatal("key allowed")
			}
			if _, err := s.SignInstallationProof(ctx, key, proof); err == nil {
				t.Fatal("proof allowed")
			}
			if called {
				t.Fatal("unverified helper received installation request")
			}
		})
	}
}

func TestInstallationHelperRejectsAmbiguousResponses(t *testing.T) {
	key, proof, _ := installationLiveFixture(t)
	valid, err := json.Marshal(map[string]any{"version": 1, "ok": true, "data": key})
	if err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{
		strings.Replace(string(valid), `"version":1`, `"version":1,"version":1`, 1),
		strings.Replace(string(valid), `"ok":true`, `"ok":true,"OK":true`, 1),
		strings.Replace(string(valid), `"algorithm":`, `"extra":null,"algorithm":`, 1),
		strings.Replace(string(valid), `"algorithm":`, `"algorithm":"wrong","algorithm":`, 1),
		strings.Replace(string(valid), key.KeyFingerprint, strings.Repeat("0", 64), 1),
		`{"version":1,"ok":true,"data":null}`, `{"version":1,"ok":true,"error":{"code":"NOT_FOUND"}}`,
		string(valid) + ` {}`, strings.Repeat(" ", maxMessageBytes+1),
	} {
		s := testStore(func(context.Context, string, []byte) ([]byte, error) { return []byte(body), nil })
		if _, err := s.EnsureInstallationKey(context.Background(), proof.Resource); !errors.Is(err, ErrUnavailable) {
			t.Fatal("ambiguous result accepted")
		}
	}
}

func TestInstallationProofRejectsWrongIntentBeforeHelper(t *testing.T) {
	key, proof, _ := installationLiveFixture(t)
	for _, change := range []func(*InstallationProof){
		func(p *InstallationProof) { p.ClientID = "other" },
		func(p *InstallationProof) { p.Resource = "http://evil.example" },
		func(p *InstallationProof) { p.State += "=" },
		func(p *InstallationProof) { p.CodeChallenge = "short" },
		func(p *InstallationProof) { p.KeyFingerprint = strings.Repeat("0", 64) },
		func(p *InstallationProof) { p.ChallengeHash = strings.ToUpper(p.ChallengeHash) },
		func(p *InstallationProof) { p.Intent = "short" },
		func(p *InstallationProof) { p.RedirectURI = "http://localhost:49152/callback" },
		func(p *InstallationProof) { p.ExpiresAt = time.Now().Add(-time.Second).Unix() },
		func(p *InstallationProof) { p.Code = strings.Repeat("c", 257) },
	} {
		p := proof
		change(&p)
		called := false
		s := testStore(func(context.Context, string, []byte) ([]byte, error) { called = true; return nil, nil })
		if _, err := s.SignInstallationProof(context.Background(), key, p); !errors.Is(err, ErrUnavailable) || called {
			t.Fatal("invalid proof reached helper")
		}
	}
}

func TestInstallationProofRefusesInvalidSignatureAndPreservesErrors(t *testing.T) {
	key, proof, _ := installationLiveFixture(t)
	for _, code := range []string{"INTERACTION_REQUIRED", "CANCELLED", "NOT_FOUND", "TIMED_OUT", "UNKNOWN"} {
		s := testStore(func(context.Context, string, []byte) ([]byte, error) {
			return json.Marshal(map[string]any{"version": 1, "ok": false, "error": map[string]string{"code": code, "message": "private-helper-detail"}})
		})
		_, err := s.SignInstallationProof(context.Background(), key, proof)
		if err == nil || strings.Contains(err.Error(), "private-helper-detail") {
			t.Fatal("unsanitized error")
		}
		if code == "TIMED_OUT" && !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal("timeout identity lost")
		}
	}
	s := testStore(func(context.Context, string, []byte) ([]byte, error) {
		return json.Marshal(map[string]any{"version": 1, "ok": true, "data": map[string]string{"algorithm": InstallationKeyAlgorithm, "signature": "AQID"}})
	})
	if _, err := s.SignInstallationProof(context.Background(), key, proof); !errors.Is(err, ErrUnavailable) {
		t.Fatal("unverified signature returned")
	}
}
