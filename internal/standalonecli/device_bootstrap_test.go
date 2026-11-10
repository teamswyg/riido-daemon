package standalonecli

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func TestDesktopBootstrapRetainsCredentialCustodyAndBindsRequests(t *testing.T) {
	var f *linkFixture
	identity := userauth.BootstrapIdentity{ClientID: "riido-cli", InstallationID: "installation", OwnerID: "owner"}
	reqID := strings.Repeat("A", 43)
	requests := 0
	mode := "pending"
	f = newLinkFixture(t, func(w http.ResponseWriter, r *http.Request) {
		body := linkRequest(t, r)
		if r.Header.Get("Authorization") == "" {
			t.Error("missing original bearer")
		}
		switch r.URL.Path {
		case "/oauth/cli/installation-transfers/device-bootstraps/identity":
			if len(body) != 0 {
				t.Error("identity request was not empty")
			}
			_ = json.NewEncoder(w).Encode(identity)
		case "/oauth/cli/installation-transfers/device-bootstraps":
			requests++
			if len(body) != 6 || body["key_fingerprint"] != identity.KeyFingerprint || body["installation_id"] != identity.InstallationID || body["request_id"] != reqID {
				t.Error("invalid admission fields")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"request_id": reqID, "state": mode, "expires_at": int64(1800000000)})
		default:
			t.Error("unexpected path")
			w.WriteHeader(http.StatusNotFound)
		}
	})
	identity.KeyFingerprint = f.binding.KeyFingerprint
	identity.Resource = f.binding.Resource
	got, err := f.service.DesktopBootstrap(t.Context(), nil)
	if err != nil || got != identity {
		t.Fatal("identity mismatch", err)
	}
	input := userauth.BootstrapRequest{BootstrapIdentity: identity, DeviceID: "dev_" + strings.Repeat("1", 32), RequestID: reqID, RecoverySecret: strings.Repeat("B", 42) + "A", NextHash: strings.Repeat("c", 64)}
	for _, state := range []string{"pending", "issued", "prepared", "complete"} {
		mode = state
		raw, _ := json.Marshal(input)
		v, err := f.service.DesktopBootstrap(t.Context(), raw)
		if err != nil {
			t.Fatal(err)
		}
		result := v.(userauth.DeviceLinkApproval)
		if result.State != state || result.RequestID != reqID {
			t.Fatal("wrong phase")
		}
	}
	for _, field := range []string{"resource", "owner_principal_id", "installation_id", "key_fingerprint", "next_secret_hash", "recovery_secret", "expected_device_id"} {
		raw, _ := json.Marshal(input)
		var fields map[string]string
		_ = json.Unmarshal(raw, &fields)
		fields[field] = "wrong"
		raw, _ = json.Marshal(fields)
		if _, err := f.service.DesktopBootstrap(t.Context(), raw); err == nil {
			t.Fatal("accepted changed " + field)
		}
	}
	if requests != 4 {
		t.Fatal("invalid binding reached admission")
	}
}
