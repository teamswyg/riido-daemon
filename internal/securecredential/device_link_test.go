package securecredential

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func deviceLinkFixture(t *testing.T) (InstallationPublicKey, DeviceLinkProof, *ecdsa.PrivateKey) {
	t.Helper()
	key, oauth, privateKey := installationLiveFixture(t)
	return key, DeviceLinkProof{Resource: oauth.Resource, ClientID: "riido-cli", InstallationID: "installation", IntentID: "intent", ChallengeHash: strings.Repeat("a", 64), KeyFingerprint: key.KeyFingerprint, ExpectedDeviceID: "device", ApprovedPrincipalID: "approved-human", ExpiresAt: time.Now().Add(5 * time.Minute).UTC().Format(time.RFC3339Nano)}, privateKey
}
func TestDeviceLinkProofMatchesIndependentCPGolden(t *testing.T) {
	p := DeviceLinkProof{Resource: "https://cli.fixture.invalid", ClientID: "riido-cli", InstallationID: "installation", IntentID: "intent", ChallengeHash: strings.Repeat("a", 64), KeyFingerprint: strings.Repeat("b", 64), ExpectedDeviceID: "device", ApprovedPrincipalID: "approved-human", ExpiresAt: "2026-10-28T01:02:03.123Z"}
	digest := sha256.Sum256(p.framedMessage())
	if hex.EncodeToString(digest[:]) != "979957be1deb56d1c5f4a68a6153ab690cc2c394d2551c3b0f1d7ed99db3a809" {
		t.Fatal("CP wire differs")
	}
	if strings.Contains(fmt.Sprintf("%v %#v", p, p), p.ApprovedPrincipalID) {
		t.Fatal("proof logging exposes binding")
	}
}
func TestDeviceLinkTypedHelperWithoutDaemonOrCredentials(t *testing.T) {
	key, proof, privateKey := deviceLinkFixture(t)
	calls := 0
	s := testStore(func(_ context.Context, _ string, raw []byte) ([]byte, error) {
		calls++
		var sent request
		if json.Unmarshal(raw, &sent) != nil || sent.Operation != "installation-device-link-proof" || sent.Key != installationStorageKey(proof.Resource) || !sent.Interactive {
			t.Fatal("wrong helper operation or scope")
		}
		value, err := base64.StdEncoding.DecodeString(sent.Value)
		if err != nil {
			t.Fatal(err)
		}
		var actual DeviceLinkProof
		if json.Unmarshal(value, &actual) != nil || actual != proof {
			t.Fatal("proof binding changed")
		}
		message, err := actual.message(key, time.Now())
		if err != nil {
			t.Fatal(err)
		}
		digest := sha256.Sum256(message)
		signature, err := ecdsa.SignASN1(rand.Reader, privateKey, digest[:])
		if err != nil {
			t.Fatal(err)
		}
		return json.Marshal(map[string]any{"version": 1, "ok": true, "data": map[string]string{"algorithm": InstallationKeyAlgorithm, "signature": base64.RawURLEncoding.EncodeToString(signature)}})
	})
	sig, err := s.SignDeviceLinkProof(context.Background(), key, proof)
	if err != nil || sig == "" || calls != 1 || s.Key != "account-key-0123456789" {
		t.Fatal("typed proof failed or credential scope changed")
	}
}
func TestDeviceLinkRejectsMalformedProofBeforeHelper(t *testing.T) {
	key, proof, _ := deviceLinkFixture(t)
	changes := map[string]func(*DeviceLinkProof){
		"resource": func(p *DeviceLinkProof) { p.Resource = "http://localhost" }, "client": func(p *DeviceLinkProof) { p.ClientID = "other" },
		"installation": func(p *DeviceLinkProof) { p.InstallationID = "../installation" }, "intent": func(p *DeviceLinkProof) { p.IntentID = " intent" },
		"device": func(p *DeviceLinkProof) { p.ExpectedDeviceID = "" }, "principal": func(p *DeviceLinkProof) { p.ApprovedPrincipalID = "\nowner" },
		"digest": func(p *DeviceLinkProof) { p.ChallengeHash = strings.Repeat("A", 64) }, "key": func(p *DeviceLinkProof) { p.KeyFingerprint = strings.Repeat("0", 64) },
		"expired":      func(p *DeviceLinkProof) { p.ExpiresAt = time.Now().Add(-time.Minute).UTC().Format(time.RFC3339Nano) },
		"noncanonical": func(p *DeviceLinkProof) { p.ExpiresAt = "2099-10-28T01:02:03.120Z" }, "offset": func(p *DeviceLinkProof) { p.ExpiresAt = "2099-10-28T01:02:03+00:00" },
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			p := proof
			change(&p)
			s := testStore(func(context.Context, string, []byte) ([]byte, error) {
				t.Fatal("invalid input reached helper")
				return nil, nil
			})
			if _, err := s.SignDeviceLinkProof(context.Background(), key, p); !errors.Is(err, ErrUnavailable) {
				t.Fatal("invalid proof allowed")
			}
		})
	}
}
func TestDeviceLinkRejectsUnverifiedHelperAndCrossBindingSignature(t *testing.T) {
	for _, mode := range []string{"noninteractive", "signature-preflight", "cross-device", "cross-intent", "unknown-response", "duplicate-response", "cancelled"} {
		t.Run(mode, func(t *testing.T) {
			key, proof, privateKey := deviceLinkFixture(t)
			calls := 0
			s := testStore(func(_ context.Context, _ string, _ []byte) ([]byte, error) {
				calls++
				changed := proof
				if mode == "cross-device" {
					changed.ExpectedDeviceID += "-other"
				}
				if mode == "cross-intent" {
					changed.IntentID += "-other"
				}
				message, err := changed.message(key, time.Now())
				if err != nil {
					t.Fatal(err)
				}
				digest := sha256.Sum256(message)
				signature, err := ecdsa.SignASN1(rand.Reader, privateKey, digest[:])
				if err != nil {
					t.Fatal(err)
				}
				data := map[string]string{"algorithm": InstallationKeyAlgorithm, "signature": base64.RawURLEncoding.EncodeToString(signature)}
				if mode == "unknown-response" {
					data["private_key"] = "must-not-exist"
				}
				raw, err := json.Marshal(map[string]any{"version": 1, "ok": true, "data": data})
				if mode == "duplicate-response" {
					raw = []byte(strings.Replace(string(raw), `"algorithm":`, `"algorithm":"ignored","algorithm":`, 1))
				}
				return raw, err
			})
			ctx := context.Background()
			if mode == "noninteractive" {
				s.Interactive = false
			}
			if mode == "signature-preflight" {
				s.verify = func(context.Context, string) error { return ErrUnavailable }
			}
			if mode == "cancelled" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			if _, err := s.SignDeviceLinkProof(ctx, key, proof); err == nil {
				t.Fatal("untrusted signature accepted")
			}
			if (mode == "noninteractive" || mode == "signature-preflight" || mode == "cancelled") && calls != 0 {
				t.Fatal("unverified call reached helper")
			}
		})
	}
}
