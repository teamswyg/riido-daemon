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
)

func TestInstallationRepairProofMatchesServerFrame(t *testing.T) {
	key, _, _ := installationGolden(t)
	proof := InstallationRepairProof{Resource: "https://cli.fixture.invalid", ClientID: "riido-cli", RequestID: strings.Repeat("A", 43), KeyFingerprint: key.KeyFingerprint, RecoveryHash: strings.Repeat("a", 64)}
	message, err := proof.message(key)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(message)
	if hex.EncodeToString(digest[:]) != "0c57a0f38699744f88d32a8aaff7b224426374a3a9b3c8399731b023a6d038de" {
		t.Fatal("repair proof differs from independent server frame")
	}
	if strings.Contains(fmt.Sprintf("%v %#v", proof, proof), proof.RecoveryHash) {
		t.Fatal("proof logging leaks recovery binding")
	}
}

func TestInstallationRepairTypedSignerRejectsCrossPurposeAndUntrustedHelper(t *testing.T) {
	for _, mode := range []string{"valid", "cross-request", "cross-recovery", "oauth-domain", "unknown-field", "duplicate-field", "noninteractive", "unverified", "cancelled", "older-helper"} {
		t.Run(mode, func(t *testing.T) {
			key, oauth, privateKey := installationLiveFixture(t)
			proof := InstallationRepairProof{Resource: oauth.Resource, ClientID: "riido-cli", RequestID: strings.Repeat("A", 43), KeyFingerprint: key.KeyFingerprint, RecoveryHash: strings.Repeat("a", 64)}
			calls := 0
			s := testStore(func(_ context.Context, _ string, raw []byte) ([]byte, error) {
				calls++
				var sent request
				if json.Unmarshal(raw, &sent) != nil || sent.Operation != "installation-repair-proof" || sent.Key != installationStorageKey(proof.Resource) || !sent.Interactive {
					t.Fatal("wrong typed signer or storage scope")
				}
				value, err := base64.StdEncoding.DecodeString(sent.Value)
				if err != nil {
					t.Fatal(err)
				}
				var actual InstallationRepairProof
				if json.Unmarshal(value, &actual) != nil || actual != proof {
					t.Fatal("repair payload changed")
				}
				if mode == "older-helper" {
					return []byte(`{"version":1,"ok":false,"error":{"code":"INVALID_REQUEST"}}`), nil
				}
				if mode == "cross-request" {
					actual.RequestID = strings.Repeat("B", 42) + "A"
				}
				if mode == "cross-recovery" {
					actual.RecoveryHash = strings.Repeat("b", 64)
				}
				message, err := actual.message(key)
				if err != nil {
					t.Fatal(err)
				}
				if mode == "oauth-domain" {
					message = []byte("riido-cli-installation-proof.v1")
				}
				digest := sha256.Sum256(message)
				signature, err := ecdsa.SignASN1(rand.Reader, privateKey, digest[:])
				if err != nil {
					t.Fatal(err)
				}
				fields := map[string]string{"algorithm": InstallationKeyAlgorithm, "signature": base64.RawURLEncoding.EncodeToString(signature)}
				if mode == "unknown-field" {
					fields["private_key"] = "not-allowed"
				}
				result, err := json.Marshal(map[string]any{"version": 1, "ok": true, "data": fields})
				if mode == "duplicate-field" {
					result = []byte(strings.Replace(string(result), `"algorithm":`, `"algorithm":"ignored","algorithm":`, 1))
				}
				return result, err
			})
			ctx := t.Context()
			if mode == "noninteractive" {
				s.Interactive = false
			}
			if mode == "unverified" {
				s.verify = func(context.Context, string) error { return ErrUnavailable }
			}
			if mode == "cancelled" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			sig, err := s.SignInstallationRepairProof(ctx, key, proof)
			if mode == "valid" {
				if err != nil || sig == "" || calls != 1 {
					t.Fatal("valid typed signer failed", err)
				}
			} else if err == nil || sig != "" {
				t.Fatal("untrusted proof accepted")
			}
			if (mode == "noninteractive" || mode == "unverified" || mode == "cancelled") && calls != 0 {
				t.Fatal("unverified request reached helper")
			}
		})
	}
}

func TestInstallationRepairRejectsInvalidInputBeforeHelper(t *testing.T) {
	key, oauth, _ := installationLiveFixture(t)
	base := InstallationRepairProof{Resource: oauth.Resource, ClientID: "riido-cli", RequestID: strings.Repeat("A", 43), KeyFingerprint: key.KeyFingerprint, RecoveryHash: strings.Repeat("a", 64)}
	for _, field := range []string{"resource", "client", "request", "recovery", "key"} {
		t.Run(field, func(t *testing.T) {
			proof := base
			switch field {
			case "resource":
				proof.Resource = "http://localhost"
			case "client":
				proof.ClientID = "other"
			case "request":
				proof.RequestID += "="
			case "recovery":
				proof.RecoveryHash = strings.Repeat("A", 64)
			case "key":
				proof.KeyFingerprint = strings.Repeat("b", 64)
			}
			s := testStore(func(context.Context, string, []byte) ([]byte, error) {
				t.Fatal("invalid proof reached helper")
				return nil, nil
			})
			if _, err := s.SignInstallationRepairProof(t.Context(), key, proof); !errors.Is(err, ErrUnavailable) {
				t.Fatal("invalid input accepted", err)
			}
		})
	}
}
