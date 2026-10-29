package saasplane

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func cliLinkTestProof() contextcli.CLIInstallationLinkProof {
	return contextcli.CLIInstallationLinkProof{Resource: "https://resource.example.test", ClientID: "riido-cli", InstallationID: "installation", IntentID: "intent", ChallengeHash: strings.Repeat("a", 64), KeyFingerprint: strings.Repeat("b", 64), ExpectedDeviceID: "device", ApprovedPrincipalID: "owner", ExpiresAt: time.Now().Add(time.Minute).UTC().Format(time.RFC3339Nano)}
}

func cliLinkTestSignature() string {
	return base64.RawURLEncoding.EncodeToString([]byte{0x30, 6, 2, 1, 1, 2, 1, 1})
}

func cliLinkTestACK(proof contextcli.CLIInstallationLinkProof) map[string]string {
	return map[string]string{"schema_version": "riido-cli-installation-link.v1", "installation_id": proof.InstallationID, "resource": proof.Resource, "client_id": proof.ClientID, "key_fingerprint": proof.KeyFingerprint, "device_id": proof.ExpectedDeviceID, "link_approved_by": proof.ApprovedPrincipalID, "device_owner_at_link": proof.ApprovedPrincipalID, "link_revision": "1", "linked_at": time.Now().UTC().Format(time.RFC3339Nano)}
}

func cliLinkTestClient(t *testing.T, handler http.Handler) (*CLIInstallationLinkClient, contextcli.CLIInstallationLinkProof) {
	t.Helper()
	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)
	proof := cliLinkTestProof()
	client, err := NewCLIInstallationLinkClient(CLIInstallationLinkConfig{BaseURL: server.URL, Resource: proof.Resource, DeviceID: proof.ExpectedDeviceID, DeviceSecret: "synthetic-device-secret", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	return client, proof
}

func TestCLILinkConfirmRejectsUnboundOrAmbiguousACK(t *testing.T) {
	for _, field := range []string{"schema_version", "installation_id", "resource", "client_id", "key_fingerprint", "device_id", "link_approved_by", "device_owner_at_link", "link_revision", "linked_at"} {
		t.Run(field, func(t *testing.T) {
			client, proof := cliLinkTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				ack := cliLinkTestACK(cliLinkTestProof())
				ack[field] = "mismatched"
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(ack)
			}))
			_, err := client.Confirm(context.Background(), proof, cliLinkTestSignature())
			requireCLILinkCode(t, err, "LINK_INVALID_RESPONSE")
		})
	}
	for name, alter := range map[string]func(string) string{
		"duplicate": func(s string) string {
			return strings.Replace(s, `"device_id":`, `"device_id":"device","device_id":`, 1)
		},
		"case alias": func(s string) string { return strings.Replace(s, `"device_id":`, `"Device_id":`, 1) },
		"unknown":    func(s string) string { return strings.TrimSuffix(s, "}") + `,"token":"synthetic-response-secret"}` },
		"null":       func(s string) string { return strings.Replace(s, `"device_id":"device"`, `"device_id":null`, 1) },
		"trailing":   func(s string) string { return s + `{}` },
		"oversized":  func(string) string { return strings.Repeat("x", (16<<10)+1) },
	} {
		t.Run(name, func(t *testing.T) {
			client, proof := cliLinkTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				raw, _ := json.Marshal(cliLinkTestACK(cliLinkTestProof()))
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, alter(string(raw)))
			}))
			_, err := client.Confirm(context.Background(), proof, cliLinkTestSignature())
			requireCLILinkCode(t, err, "LINK_INVALID_RESPONSE")
		})
	}
}

func TestCLILinkConfirmRejectsAuthorityBeforeNetwork(t *testing.T) {
	var calls atomic.Int32
	client, original := cliLinkTestClient(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls.Add(1) }))
	for name, change := range map[string]func(*contextcli.CLIInstallationLinkProof){
		"resource": func(p *contextcli.CLIInstallationLinkProof) { p.Resource = "https://other.test" },
		"device":   func(p *contextcli.CLIInstallationLinkProof) { p.ExpectedDeviceID = "other" },
		"client":   func(p *contextcli.CLIInstallationLinkProof) { p.ClientID = "other" },
		"expired": func(p *contextcli.CLIInstallationLinkProof) {
			p.ExpiresAt = time.Now().Add(-time.Minute).UTC().Format(time.RFC3339Nano)
		},
		"path":        func(p *contextcli.CLIInstallationLinkProof) { p.IntentID = "intent/other" },
		"hash":        func(p *contextcli.CLIInstallationLinkProof) { p.ChallengeHash = strings.Repeat("A", 64) },
		"fingerprint": func(p *contextcli.CLIInstallationLinkProof) { p.KeyFingerprint = "bad" },
		"principal":   func(p *contextcli.CLIInstallationLinkProof) { p.ApprovedPrincipalID = "" },
	} {
		t.Run(name, func(t *testing.T) {
			proof := original
			change(&proof)
			_, err := client.Confirm(context.Background(), proof, cliLinkTestSignature())
			requireCLILinkCode(t, err, "LINK_INVALID_REQUEST")
		})
	}
	for _, signature := range []string{"", "synthetic-secret", cliLinkTestSignature() + "=", base64.RawURLEncoding.EncodeToString([]byte{0x30, 6, 2, 1, 0, 2, 1, 1})} {
		_, err := client.Confirm(context.Background(), original, signature)
		requireCLILinkCode(t, err, "LINK_INVALID_REQUEST")
	}
	if calls.Load() != 0 {
		t.Fatal("invalid proof reached CP")
	}
}

func TestCLILinkConfigNeverFallsBackToAnotherAuthority(t *testing.T) {
	for name, change := range map[string]func(*CLIInstallationLinkConfig){
		"no resource":      func(c *CLIInstallationLinkConfig) { c.Resource = "" },
		"resource path":    func(c *CLIInstallationLinkConfig) { c.Resource += "/" },
		"resource query":   func(c *CLIInstallationLinkConfig) { c.Resource += "?" },
		"insecure CP":      func(c *CLIInstallationLinkConfig) { c.BaseURL = "http://cp.test" },
		"CP userinfo":      func(c *CLIInstallationLinkConfig) { c.BaseURL = "https://user:secret@cp.test" },
		"missing device":   func(c *CLIInstallationLinkConfig) { c.DeviceID = "" },
		"missing secret":   func(c *CLIInstallationLinkConfig) { c.DeviceSecret = "" },
		"missing coverage": func(c *CLIInstallationLinkConfig) { c.DeviceID = "dcov_device" },
	} {
		t.Run(name, func(t *testing.T) {
			cfg := CLIInstallationLinkConfig{BaseURL: "https://cp.test", Resource: "https://resource.test", DeviceID: "device", DeviceSecret: "synthetic-secret"}
			change(&cfg)
			_, err := NewCLIInstallationLinkClient(cfg)
			requireCLILinkCode(t, err, "LINK_UNSUPPORTED")
		})
	}
}

func TestCLILinkConfirmDoesNotRedirectOrReflectErrors(t *testing.T) {
	var redirected atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { redirected.Add(1) }))
	defer target.Close()
	for status, code := range map[int]string{302: "LINK_INVALID_RESPONSE", 400: "LINK_INVALID_REQUEST", 401: "LINK_PERMISSION_DENIED", 403: "LINK_PERMISSION_DENIED", 404: "LINK_UNSUPPORTED", 405: "LINK_UNSUPPORTED", 408: "LINK_UNAVAILABLE", 409: "LINK_CONFLICT", 410: "LINK_EXPIRED", 429: "LINK_UNAVAILABLE", 503: "LINK_UNAVAILABLE"} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			client, proof := cliLinkTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Location", target.URL)
				w.WriteHeader(status)
				_, _ = io.WriteString(w, "synthetic-server-secret")
			}))
			_, err := client.Confirm(context.Background(), proof, cliLinkTestSignature())
			requireCLILinkCode(t, err, code)
		})
	}
	if redirected.Load() != 0 {
		t.Fatal("device transport followed redirect")
	}
	client, proof := cliLinkTestClient(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := client.Confirm(ctx, proof, cliLinkTestSignature())
	requireCLILinkCode(t, err, "LINK_UNAVAILABLE")
}

func requireCLILinkCode(t *testing.T, err error, code string) {
	t.Helper()
	var typed *contextcli.CLIInstallationLinkError
	if !errors.As(err, &typed) || typed.Code != code || err.Error() != code {
		t.Fatalf("expected %s, received redacted error class %T", code, err)
	}
}
