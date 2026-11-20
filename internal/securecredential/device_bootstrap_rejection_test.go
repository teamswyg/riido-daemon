package securecredential

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestDeviceBootstrapRejectsInvalidIntentBeforeHelper(t *testing.T) {
	changes := map[string]func(*DeviceBootstrapProof){
		"phase":                     func(p *DeviceBootstrapProof) { p.Phase = "link" },
		"transition":                func(p *DeviceBootstrapProof) { p.TransitionID = "not-a-transition" },
		"intent":                    func(p *DeviceBootstrapProof) { p.IntentID = "legacy-intent" },
		"prepare-with-previous":     func(p *DeviceBootstrapProof) { p.PreviousBootDigest = strings.Repeat("a", 64) },
		"activate-without-previous": func(p *DeviceBootstrapProof) { p.Phase = "activate" },
		"expired": func(p *DeviceBootstrapProof) {
			p.ExpiresAt = time.Now().Add(-time.Minute).UTC().Format(time.RFC3339Nano)
		},
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			key, proof, _ := bootstrapFixture(t)
			change(&proof)
			s := testStore(func(context.Context, string, []byte) ([]byte, error) {
				t.Fatal("invalid proof reached helper")
				return nil, nil
			})
			if _, err := s.SignDeviceBootstrapProof(context.Background(), key, proof); err == nil {
				t.Fatal("invalid proof accepted")
			}
		})
	}
}

func TestDeviceBootstrapRejectsUnverifiedWitness(t *testing.T) {
	for _, mode := range []string{"same-boot", "cross-boot", "cross-device", "cross-transition", "unknown-field", "duplicate-field", "wrong-algorithm", "noninteractive", "unverified-helper", "cancelled"} {
		t.Run(mode, func(t *testing.T) {
			key, proof, private := bootstrapFixture(t)
			proof.Phase, proof.PreviousBootDigest = "activate", strings.Repeat("a", 64)
			calls := 0
			s := testStore(func(context.Context, string, []byte) ([]byte, error) {
				calls++
				changed := proof
				if mode == "cross-device" {
					changed.ExpectedDeviceID += "-other"
				}
				if mode == "cross-transition" {
					changed.TransitionID = base64.RawURLEncoding.EncodeToString([]byte(strings.Repeat("x", 32)))
				}
				raw := bootstrapResponse(t, key, private, changed, strings.Repeat("b", 64))
				var response map[string]any
				_ = json.Unmarshal(raw, &response)
				data := response["data"].(map[string]any)
				switch mode {
				case "same-boot":
					data["boot_digest"] = proof.PreviousBootDigest
				case "cross-boot":
					data["boot_digest"] = strings.Repeat("c", 64)
				case "unknown-field":
					data["boot_uuid"] = "forbidden"
				case "wrong-algorithm":
					data["algorithm"] = "untrusted"
				}
				out, err := json.Marshal(response)
				if mode == "duplicate-field" {
					out = []byte(strings.Replace(string(out), `"boot_digest":`, `"boot_digest":"ignored","boot_digest":`, 1))
				}
				return out, err
			})
			ctx := context.Background()
			if mode == "noninteractive" {
				s.Interactive = false
			}
			if mode == "unverified-helper" {
				s.verify = func(context.Context, string) error { return ErrUnavailable }
			}
			if mode == "cancelled" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			if _, err := s.SignDeviceBootstrapProof(ctx, key, proof); err == nil {
				t.Fatal("unverified witness accepted")
			}
			if (mode == "noninteractive" || mode == "unverified-helper" || mode == "cancelled") && calls != 0 {
				t.Fatal("untrusted invocation reached helper")
			}
		})
	}
}
