package securecredential

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func transferFixture(t *testing.T) (InstallationPublicKey, DeviceTransferProof, *ecdsa.PrivateKey) {
	t.Helper()
	key, bootstrap, private := bootstrapFixture(t)
	return key, DeviceTransferProof{DeviceLinkProof: bootstrap.DeviceLinkProof, PreviousDeviceID: "source-device", PreviousLinkRevision: "2"}, private
}

func transferResponse(t *testing.T, private *ecdsa.PrivateKey, message []byte) []byte {
	t.Helper()
	digest := sha256.Sum256(message)
	sig, err := ecdsa.SignASN1(rand.Reader, private, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(map[string]any{"version": 1, "ok": true, "data": map[string]string{"algorithm": InstallationKeyAlgorithm, "signature": base64.RawURLEncoding.EncodeToString(sig)}})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestDeviceTransferBindsExistingSourceWithoutCredentialChanges(t *testing.T) {
	key, proof, private := transferFixture(t)
	calls := 0
	s := testStore(func(_ context.Context, _ string, raw []byte) ([]byte, error) {
		calls++
		var sent request
		if json.Unmarshal(raw, &sent) != nil || sent.Operation != "installation-device-transfer-proof" || !sent.Interactive || sent.Key != installationStorageKey(proof.Resource) {
			t.Fatal("incorrect transfer operation or scope")
		}
		value, err := base64.StdEncoding.DecodeString(sent.Value)
		if err != nil {
			t.Fatal(err)
		}
		var actual DeviceTransferProof
		if json.Unmarshal(value, &actual) != nil || actual != proof {
			t.Fatal("changed source binding")
		}
		return transferResponse(t, private, proof.framedMessage()), nil
	})
	if sig, err := s.SignDeviceTransferProof(context.Background(), key, proof); err != nil || sig == "" || calls != 1 || s.Key != "account-key-0123456789" {
		t.Fatal("transfer failed or changed credential scope")
	}
}

func TestDeviceTransferRejectsMalformedSourceBeforeHelper(t *testing.T) {
	for _, revision := range []string{"", "0", "01", "+1", "-1", "18446744073709551615", "18446744073709551616"} {
		t.Run("revision="+revision, func(t *testing.T) {
			key, proof, _ := transferFixture(t)
			proof.PreviousLinkRevision = revision
			s := testStore(func(context.Context, string, []byte) ([]byte, error) {
				t.Fatal("invalid source reached helper")
				return nil, nil
			})
			if _, err := s.SignDeviceTransferProof(context.Background(), key, proof); err == nil {
				t.Fatal("invalid revision accepted")
			}
		})
	}
	for _, source := range []string{"", "device", "../device", " source", "device\n"} {
		key, proof, _ := transferFixture(t)
		proof.PreviousDeviceID = source
		if _, err := proof.message(key, time.Now()); err == nil {
			t.Fatal("invalid source accepted")
		}
	}
}

func TestDeviceTransferRejectsInitialLinkAndChangedSourceSignature(t *testing.T) {
	for _, mode := range []string{"initial-link", "source", "revision", "target", "unknown-response", "duplicate-response", "noninteractive", "untrusted"} {
		t.Run(mode, func(t *testing.T) {
			key, proof, private := transferFixture(t)
			calls := 0
			s := testStore(func(context.Context, string, []byte) ([]byte, error) {
				calls++
				changed := proof
				if mode == "source" {
					changed.PreviousDeviceID += "-other"
				}
				if mode == "revision" {
					changed.PreviousLinkRevision = "3"
				}
				if mode == "target" {
					changed.ExpectedDeviceID += "-other"
				}
				message := changed.framedMessage()
				if mode == "initial-link" {
					message = proof.DeviceLinkProof.framedMessage()
				}
				raw := transferResponse(t, private, message)
				if mode == "unknown-response" {
					raw = []byte(strings.Replace(string(raw), `"algorithm":`, `"private_key":"forbidden","algorithm":`, 1))
				}
				if mode == "duplicate-response" {
					raw = []byte(strings.Replace(string(raw), `"algorithm":`, `"algorithm":"ignored","algorithm":`, 1))
				}
				return raw, nil
			})
			if mode == "noninteractive" {
				s.Interactive = false
			}
			if mode == "untrusted" {
				s.verify = func(context.Context, string) error { return ErrUnavailable }
			}
			if _, err := s.SignDeviceTransferProof(context.Background(), key, proof); err == nil {
				t.Fatal("untrusted transfer accepted")
			}
			if (mode == "noninteractive" || mode == "untrusted") && calls != 0 {
				t.Fatal("unverified caller reached helper")
			}
		})
	}
}
