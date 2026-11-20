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

func TestInstallationTransferRepairProofMatchesServerFrame(t *testing.T) {
	key, _, _ := installationGolden(t)
	proof := InstallationTransferRepairProof{Resource: "https://cli.fixture.invalid", ClientID: "riido-cli", RequestID: strings.Repeat("A", 43), OriginalRequestID: strings.Repeat("B", 42) + "A", KeyFingerprint: key.KeyFingerprint, RecoveryHash: strings.Repeat("a", 64)}
	message, err := proof.message(key)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(message)
	if hex.EncodeToString(digest[:]) != "0a890fbf1ba54cc46de8773baf57950128f39b9652ec99be4d6798860a584552" {
		t.Fatal("repair proof differs from independent server frame")
	}
	if strings.Contains(fmt.Sprintf("%v %#v", proof, proof), proof.RecoveryHash) {
		t.Fatal("proof logging leaks recovery binding")
	}
}

func TestInstallationTransferRepairTypedSignerRejectsCrossPurposeAndUntrustedHelper(t *testing.T) {
	for _, mode := range []string{"valid", "cross-request", "cross-original", "initial-repair", "cross-recovery", "oauth-domain", "unknown-field", "duplicate-field", "noninteractive", "unverified", "cancelled", "older-helper"} {
		t.Run(mode, func(t *testing.T) {
			key, oauth, privateKey := installationLiveFixture(t)
			proof := InstallationTransferRepairProof{Resource: oauth.Resource, ClientID: "riido-cli", RequestID: strings.Repeat("A", 43), OriginalRequestID: strings.Repeat("B", 42) + "A", KeyFingerprint: key.KeyFingerprint, RecoveryHash: strings.Repeat("a", 64)}
			calls := 0
			s := testStore(func(_ context.Context, _ string, raw []byte) ([]byte, error) {
				calls++
				var sent request
				if json.Unmarshal(raw, &sent) != nil || sent.Operation != "installation-transfer-repair-proof" || sent.Key != installationStorageKey(proof.Resource) || !sent.Interactive {
					t.Fatal("wrong typed signer or storage scope")
				}
				value, err := base64.StdEncoding.DecodeString(sent.Value)
				if err != nil {
					t.Fatal(err)
				}
				var actual InstallationTransferRepairProof
				if json.Unmarshal(value, &actual) != nil || actual != proof {
					t.Fatal("repair payload changed")
				}
				if mode == "older-helper" {
					return []byte(`{"version":1,"ok":false,"error":{"code":"INVALID_REQUEST"}}`), nil
				}
				if mode == "cross-request" {
					actual.RequestID = strings.Repeat("C", 42) + "A"
				}
				if mode == "cross-original" {
					actual.OriginalRequestID = strings.Repeat("D", 42) + "A"
				}
				if mode == "cross-recovery" {
					actual.RecoveryHash = strings.Repeat("b", 64)
				}
				message, err := actual.message(key)
				if err != nil {
					t.Fatal(err)
				}
				if mode == "initial-repair" {
					initial := InstallationRepairProof{Resource: actual.Resource, ClientID: actual.ClientID, RequestID: actual.RequestID, KeyFingerprint: actual.KeyFingerprint, RecoveryHash: actual.RecoveryHash}
					message, err = initial.message(key)
					if err != nil {
						t.Fatal(err)
					}
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
			sig, err := s.SignInstallationTransferRepairProof(ctx, key, proof)
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

func TestInstallationTransferRepairRejectsInvalidInputBeforeHelper(t *testing.T) {
	key, oauth, _ := installationLiveFixture(t)
	base := InstallationTransferRepairProof{Resource: oauth.Resource, ClientID: "riido-cli", RequestID: strings.Repeat("A", 43), OriginalRequestID: strings.Repeat("B", 42) + "A", KeyFingerprint: key.KeyFingerprint, RecoveryHash: strings.Repeat("a", 64)}
	for _, field := range []string{"resource", "client", "request", "original", "same-request", "recovery", "key"} {
		t.Run(field, func(t *testing.T) {
			proof := base
			switch field {
			case "resource":
				proof.Resource = "http://localhost"
			case "client":
				proof.ClientID = "other"
			case "request":
				proof.RequestID += "="
			case "original":
				proof.OriginalRequestID += "="
			case "same-request":
				proof.OriginalRequestID = proof.RequestID
			case "recovery":
				proof.RecoveryHash = strings.Repeat("A", 64)
			case "key":
				proof.KeyFingerprint = strings.Repeat("b", 64)
			}
			s := testStore(func(context.Context, string, []byte) ([]byte, error) {
				t.Fatal("invalid proof reached helper")
				return nil, nil
			})
			if _, err := s.SignInstallationTransferRepairProof(t.Context(), key, proof); !errors.Is(err, ErrUnavailable) {
				t.Fatal("invalid input accepted", err)
			}
		})
	}
}
