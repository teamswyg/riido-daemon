package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func TestCLILinkIPCConfirmUsesFixedDeviceTransport(t *testing.T) {
	var calls atomic.Int32
	proof := contextcli.CLIInstallationLinkProof{Resource: "https://monolith.example.test", ClientID: "riido-cli", InstallationID: "install", IntentID: "intent", ChallengeHash: strings.Repeat("a", 64), KeyFingerprint: strings.Repeat("b", 64), ExpectedDeviceID: "fixture-device", ApprovedPrincipalID: "owner", ExpiresAt: time.Now().Add(time.Minute).UTC().Format(time.RFC3339Nano)}
	signature := base64.RawURLEncoding.EncodeToString([]byte{0x30, 6, 2, 1, 1, 2, 1, 1})
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Method != http.MethodPost || r.URL.Path != "/v1/daemon/cli-installation-links/intent/confirm" || r.URL.RawQuery != "" || r.Header.Get("X-Riido-Device-Id") != proof.ExpectedDeviceID || r.Header.Get("X-Riido-Device-Secret") != "synthetic-device-secret" || r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
			t.Error("unexpected device transport authority")
		}
		var body map[string]string
		if json.NewDecoder(r.Body).Decode(&body) != nil || len(body) != 2 || body["schema_version"] != "riido-cli-installation-link-confirm.v1" || body["signature"] != signature {
			t.Error("unexpected device transport body")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"schema_version": "riido-cli-installation-link.v1", "installation_id": proof.InstallationID, "resource": proof.Resource, "client_id": proof.ClientID, "key_fingerprint": proof.KeyFingerprint, "device_id": proof.ExpectedDeviceID, "link_approved_by": proof.ApprovedPrincipalID, "device_owner_at_link": proof.ApprovedPrincipalID, "link_revision": "1", "linked_at": time.Now().UTC().Format(time.RFC3339Nano)})
	}))
	defer server.Close()
	settings := daemonSettings{CLIAPIResource: proof.Resource, DeviceID: proof.ExpectedDeviceID, DeviceSecret: "synthetic-device-secret", SaaSURL: server.URL, cliLinkHTTPClient: server.Client()}
	request := contextcli.CLIInstallationLinkRequest{Method: contextcli.IPCMethodCLIInstallationLinkConfirm, SchemaVersion: contextcli.CLIInstallationLinkIPCSchema, Proof: &proof, Signature: signature}
	raw, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	response := cliLinkIPCFixture(t, settings, string(raw))
	var reply contextcli.CLIInstallationLinkReply
	if json.Unmarshal(response, &reply) != nil || !reply.OK || reply.Data == nil || !reply.Data.Confirmed || reply.Data.IntentID != proof.IntentID || calls.Load() != 1 {
		t.Fatal("fixed confirmation was not acknowledged")
	}
	if strings.Contains(string(response), signature) || strings.Contains(string(response), settings.DeviceSecret) {
		t.Fatal("confirmation exposed protected material")
	}
	// Unknown capability/URL fields must not reach CP or fall through to user auth.
	bad := strings.TrimSuffix(string(raw), "}") + `,"run_capability":"synthetic-secret"}`
	response = cliLinkIPCFixture(t, settings, bad)
	reply = contextcli.CLIInstallationLinkReply{}
	if json.Unmarshal(response, &reply) != nil || reply.OK || calls.Load() != 1 || strings.Contains(string(response), "synthetic-secret") {
		t.Fatal("unexpected confirmation fields were accepted or reflected")
	}
}
