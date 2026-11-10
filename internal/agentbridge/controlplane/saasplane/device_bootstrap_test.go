package saasplane

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
)

type bootstrapWitnessFixture struct {
	key, boot string
	count     int
	verify    func(securecredential.DeviceBootstrapProof)
}

func (f *bootstrapWitnessFixture) ExistingInstallationKey(context.Context, string) (securecredential.InstallationPublicKey, error) {
	return securecredential.InstallationPublicKey{KeyFingerprint: f.key}, nil
}

func (f *bootstrapWitnessFixture) SignDeviceBootstrapProof(_ context.Context, _ securecredential.InstallationPublicKey, p securecredential.DeviceBootstrapProof) (securecredential.DeviceBootstrapSignature, error) {
	f.count++
	f.verify(p)
	return securecredential.DeviceBootstrapSignature{BootDigest: f.boot, Signature: "MAYCAQECAQE"}, nil
}

func TestNativeBootstrapHTTPRequiresBoundWitnessAndRecoversACK(t *testing.T) {
	for _, scenario := range []string{"prepare", "lost-prepare", "prepared-retry", "activate", "renew-prepared", "same-boot", "wrong-owner", "expired-issued", "redirect", "wrong-completion"} {
		t.Run(scenario, func(t *testing.T) {
			id := strings.Repeat("A", 43)
			q := bootstrapRequest{resource: "https://monolith.api.riido.io", installation: "fixture-install", key: strings.Repeat("a", 64), device: "dev_" + strings.Repeat("2", 32), secret: "test-native", owner: "fixture-owner", id: id, nextHash: strings.Repeat("3", 64)}
			state := "issued"
			action := "prepare"
			oldBoot := strings.Repeat("4", 64)
			witness := &bootstrapWitnessFixture{key: q.key, boot: oldBoot}
			exp := time.Now().UTC().Add(time.Minute)
			if scenario == "prepared-retry" || scenario == "activate" || scenario == "renew-prepared" || scenario == "same-boot" || scenario == "wrong-completion" {
				state = "prepared"
			}
			if state == "prepared" && scenario != "prepared-retry" {
				action = "activate"
				witness.boot = strings.Repeat("5", 64)
			}
			if scenario == "same-boot" {
				witness.boot = oldBoot
			}
			if scenario == "expired-issued" || scenario == "renew-prepared" {
				exp = time.Now().UTC().Add(-time.Minute)
			}
			witness.verify = func(p securecredential.DeviceBootstrapProof) {
				if p.Resource != q.resource || p.TransitionID != q.id || p.Phase != action || p.ExpectedDeviceID != q.device || p.ApprovedPrincipalID != q.owner || p.IntentID != id || p.ChallengeHash != q.nextHash || p.KeyFingerprint != q.key {
					t.Error("signed unrelated challenge")
				}
			}
			var ops []string
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				op := strings.TrimPrefix(r.URL.Path, "/v1/daemon/device-bootstraps/"+id+"/")
				ops = append(ops, op)
				if r.Method != http.MethodPost || r.URL.RawQuery != "" || r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" || r.Header.Get("X-Riido-Device-Secret") != q.secret || r.Header.Get("X-Riido-Device-Id") != q.device {
					t.Error("unexpected credentials or endpoint")
					w.WriteHeader(http.StatusForbidden)
					return
				}
				var body map[string]string
				if json.NewDecoder(r.Body).Decode(&body) != nil || body["schema_version"] != deviceBootstrapSchema {
					t.Error("invalid body")
				}
				switch op {
				case "status":
					if len(body) != 1 {
						t.Error("injected application data")
					}
				case "renew":
					if scenario != "renew-prepared" || len(body) != 1 {
						t.Error("unexpected renewal")
					}
					exp = time.Now().UTC().Add(time.Minute)
				case "prepare", "activate":
					if body["intent_id"] != id || body["signature"] != "MAYCAQECAQE" || body["boot_digest"] != witness.boot || len(body) != 4 {
						t.Error("invalid witness body")
					}
					if op == "prepare" {
						state = "prepared"
					} else {
						state = "complete"
					}
				default:
					t.Error("unknown operation")
				}
				if scenario == "redirect" {
					http.Redirect(w, r, "/forbidden", http.StatusTemporaryRedirect)
					return
				}
				if scenario == "lost-prepare" && op == "prepare" {
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}
				f := map[string]string{"schema_version": deviceBootstrapSchema, "resource": q.resource, "client_id": "riido-cli", "installation_id": q.installation, "key_fingerprint": q.key, "transition_id": q.id, "device_id": q.device, "owner_principal_id": q.owner, "state": state, "intent_id": id, "challenge_digest": q.nextHash, "expires_at": exp.Format(time.RFC3339Nano)}
				if scenario == "wrong-owner" {
					f["owner_principal_id"] = "other"
				}
				if state != "issued" {
					f["previous_boot_digest"] = oldBoot
				}
				if state == "complete" {
					h := sha256.Sum256([]byte(strings.Join([]string{deviceBootstrapSchema, q.id, q.device, q.key, q.nextHash}, "\x00")))
					f["execution_coverage_generation"] = "covb_" + hex.EncodeToString(h[:])
					f["credential_generation"] = "cred_" + q.nextHash
					f["credential_issued_at"] = exp.Format(time.RFC3339Nano)
					f["completed_at"] = exp.Format(time.RFC3339Nano)
					if scenario == "wrong-completion" {
						f["execution_coverage_generation"] = "covb_" + q.key
					}
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(f)
			}))
			defer srv.Close()
			q.origin = srv.URL
			out, err := q.advance(t.Context(), action, witness, srv.Client())
			switch scenario {
			case "prepare", "prepared-retry":
				if err != nil || out != "prepared" {
					t.Fatal(out, err)
				}
			case "activate", "renew-prepared":
				if err != nil || out != "complete" {
					t.Fatal(out, err)
				}
			case "lost-prepare":
				if err == nil {
					t.Fatal("lost ACK was accepted")
				}
				out, err = q.advance(t.Context(), action, witness, srv.Client())
				if err != nil || out != "prepared" || witness.count != 1 {
					t.Fatal("did not recover same fence")
				}
			default:
				if err == nil || out != "" {
					t.Fatal("unsafe result accepted")
				}
			}
			if (scenario == "expired-issued" || scenario == "wrong-owner" || scenario == "redirect" || scenario == "prepared-retry") && (len(ops) != 1 || witness.count != 0) {
				t.Fatal("unexpected side effect", ops)
			}
			if scenario == "same-boot" && len(ops) != 1 {
				t.Fatal("sent activation from original boot")
			}
		})
	}
}

func TestNativeBootstrapInputRejectsInjectedEndpointAndProof(t *testing.T) {
	request := map[string]string{"resource": "https://monolith.api.riido.io", "control_plane_origin": "https://staging.ai-api.riido.io", "installation_id": "i", "key_fingerprint": strings.Repeat("b", 64), "device_id": "dev_" + strings.Repeat("a", 32), "device_secret": "synthetic-native", "owner_principal_id": "human", "transition_id": strings.Repeat("A", 43), "next_secret_hash": strings.Repeat("c", 64)}
	valid, _ := json.Marshal(request)
	q, err := parseBootstrapInput(valid)
	if err != nil || q.origin != "https://staging.ai-api.riido.io" {
		t.Fatal("trusted resource mapping failed")
	}
	for _, field := range []string{"resource", "control_plane_origin", "device_id", "key_fingerprint", "transition_id", "endpoint", "intent_id", "signature", "boot_digest"} {
		altered := maps.Clone(request)
		altered[field] = "https://outside.invalid"
		data, _ := json.Marshal(altered)
		if _, err := parseBootstrapInput(data); err == nil {
			t.Fatal("accepted " + field)
		}
	}
}
