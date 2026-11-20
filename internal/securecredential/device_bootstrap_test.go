package securecredential

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func bootstrapFixture(t *testing.T) (InstallationPublicKey, DeviceBootstrapProof, *ecdsa.PrivateKey) {
	t.Helper()
	key, link, private := deviceLinkFixture(t)
	link.IntentID = base64.RawURLEncoding.EncodeToString([]byte(strings.Repeat("i", 32)))
	return key, DeviceBootstrapProof{
		DeviceLinkProof: link, Phase: "prepare",
		TransitionID: base64.RawURLEncoding.EncodeToString([]byte(strings.Repeat("t", 32))),
	}, private
}

func bootstrapResponse(t *testing.T, key InstallationPublicKey, private *ecdsa.PrivateKey, proof DeviceBootstrapProof, boot string) []byte {
	t.Helper()
	message, err := proof.message(key, boot, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(message)
	sig, err := ecdsa.SignASN1(rand.Reader, private, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(map[string]any{"version": 1, "ok": true, "data": map[string]string{
		"algorithm": InstallationKeyAlgorithm, "signature": base64.RawURLEncoding.EncodeToString(sig), "boot_digest": boot,
	}})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestDeviceBootstrapSignsFixedIntentWithoutChangingCredentials(t *testing.T) {
	for _, phase := range []string{"prepare", "activate"} {
		t.Run(phase, func(t *testing.T) {
			key, proof, private := bootstrapFixture(t)
			proof.Phase = phase
			if phase == "activate" {
				proof.PreviousBootDigest = strings.Repeat("a", 64)
			}
			calls := 0
			s := testStore(func(_ context.Context, _ string, raw []byte) ([]byte, error) {
				calls++
				var sent request
				if json.Unmarshal(raw, &sent) != nil || sent.Operation != "installation-device-bootstrap-proof" || !sent.Interactive || sent.Key != installationStorageKey(proof.Resource) {
					t.Fatal("wrong signing operation or scope")
				}
				payload, err := base64.StdEncoding.DecodeString(sent.Value)
				if err != nil {
					t.Fatal(err)
				}
				var actual DeviceBootstrapProof
				if json.Unmarshal(payload, &actual) != nil || actual != proof {
					t.Fatal("changed intent")
				}
				var fields map[string]any
				if json.Unmarshal(payload, &fields) != nil || len(fields) != 12 || fields["boot_digest"] != nil {
					t.Fatal("caller supplied current boot witness")
				}
				return bootstrapResponse(t, key, private, proof, strings.Repeat("b", 64)), nil
			})
			result, err := s.SignDeviceBootstrapProof(context.Background(), key, proof)
			if err != nil || result.BootDigest != strings.Repeat("b", 64) || calls != 1 || s.Key != "account-key-0123456789" {
				t.Fatal("fixed intent or credential scope failed")
			}
			for _, value := range []any{proof, result} {
				if !strings.Contains(fmt.Sprintf("%v %#v", value, value), "redacted") {
					t.Fatal("unsafe proof logging")
				}
			}
		})
	}
}
