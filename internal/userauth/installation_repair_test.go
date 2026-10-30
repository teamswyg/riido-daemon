package userauth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func repairFixtureRequest() InstallationRepairRequest {
	r := linkFixtureRequest()
	return InstallationRepairRequest{RequestID: r.RequestID, RecoveryHash: strings.Repeat("a", 64), Key: fixtureInstallationKey(), Signature: "MAYCAQECAQE"}
}

func repairFixtureRecovery() InstallationRepairRecoveryRequest {
	r := linkFixtureRequest()
	return InstallationRepairRecoveryRequest{RequestID: r.RequestID, RecoverySecret: r.RecoverySecret, KeyFingerprint: fixtureInstallationKey().KeyFingerprint}
}

func TestInstallationRepairTransportUsesOnlyTypedAuthority(t *testing.T) {
	input := repairFixtureRequest()
	for _, create := range []bool{true, false} {
		t.Run(map[bool]string{true: "create", false: "recover"}[create], func(t *testing.T) {
			calls := 0
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != http.MethodPost || r.URL.RawQuery != "" || r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" || r.Header.Get("Origin") != "" {
					t.Error("ambient authority or wrong request")
				}
				var fields map[string]string
				if json.NewDecoder(r.Body).Decode(&fields) != nil {
					t.Fatal("bad payload")
				}
				if create {
					if r.URL.Path != "/oauth/cli/installation-repairs" || len(fields) != 7 || fields["recovery_hash"] != input.RecoveryHash || fields["public_key"] != input.Key.PublicKey || fields["request_id"] != input.RequestID || fields["signature"] != input.Signature {
						t.Error("wrong create custody")
					}
					linkReply(w, 201, map[string]any{"request_id": input.RequestID, "state": "approved", "expires_at": 1, "recovery_until": 2})
				} else {
					if r.URL.Path != "/oauth/cli/installation-repairs/recover" || len(fields) != 2 || fields["request_id"] != input.RequestID || fields["recovery_secret"] != repairFixtureRecovery().RecoverySecret {
						t.Error("wrong recovery custody")
					}
					linkReply(w, 200, map[string]string{"request_id": input.RequestID, "state": "linked"})
				}
			}))
			defer server.Close()
			client := linkFixtureClient(t, server)
			if create {
				result, err := client.CreateInstallationRepair(context.Background(), input)
				if err != nil || result.ExpiresAt != 1 {
					t.Fatal("create changed deadline", err)
				}
			} else {
				result, err := client.RecoverInstallationRepair(context.Background(), repairFixtureRecovery())
				if err != nil || result.State != "linked" {
					t.Fatal("recover failed", err)
				}
			}
			if calls != 1 {
				t.Fatal("unexpected extra call")
			}
		})
	}
}

func TestInstallationRepairRecoveryStrictStatesAndErrors(t *testing.T) {
	input := repairFixtureRecovery()
	for _, state := range []string{"awaiting_approval", "preparing", "approval_expired", "repair_expired", "linked"} {
		t.Run(state, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				linkReply(w, 200, map[string]string{"request_id": input.RequestID, "state": state})
			}))
			defer server.Close()
			result, err := linkFixtureClient(t, server).RecoverInstallationRepair(context.Background(), input)
			if err != nil || result.State != state {
				t.Fatal("state rejected", err)
			}
		})
	}
	for _, test := range []struct {
		status     int
		code, want string
	}{{400, "invalid_installation_repair", "REPAIR_NOT_FOUND"}, {400, "invalid_request", "REPAIR_REQUEST_REJECTED"}, {403, "access_denied", "REPAIR_REQUEST_REJECTED"}, {409, "installation_policy_unknown", "REPAIR_POLICY_UNKNOWN"}, {429, "rate_limited", "REPAIR_RATE_LIMITED"}, {503, "unavailable", "REPAIR_HTTP_ERROR"}} {
		t.Run(test.code, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				linkReply(w, test.status, map[string]string{"error": test.code, "error_description": "secret-provider-data"})
			}))
			defer server.Close()
			_, err := linkFixtureClient(t, server).RecoverInstallationRepair(context.Background(), input)
			var failure *InstallationRepairError
			if !errors.As(err, &failure) || failure.Code != test.want || failure.StatusCode != test.status || strings.Contains(err.Error(), "secret-provider-data") || errors.Is(err, ErrReauthRequired) {
				t.Fatal("unsafe repair error", err)
			}
		})
	}
}

func TestInstallationRepairRejectsMalformedAndRedirectedRecovery(t *testing.T) {
	input := repairFixtureRecovery()
	for _, body := range []string{`{"request_id":"` + input.RequestID + `","state":"linked","state":"linked"}`, `{"request_id":"other","state":"linked"}`, `{"request_id":"` + input.RequestID + `","state":"linked","proof":null}`, `{"request_id":"` + input.RequestID + `","state":"proof_expired"}`, strings.Repeat("x", 4097)} {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, body) }))
		_, err := linkFixtureClient(t, server).RecoverInstallationRepair(context.Background(), input)
		server.Close()
		var failure *InstallationRepairError
		if !errors.As(err, &failure) || failure.Code != "REPAIR_INVALID_RESPONSE" {
			t.Fatal("bad shape accepted", err)
		}
	}
	calls := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Location", "/oauth/token")
		w.WriteHeader(http.StatusTemporaryRedirect)
	}))
	defer server.Close()
	if _, err := linkFixtureClient(t, server).RecoverInstallationRepair(context.Background(), input); err == nil || calls != 1 {
		t.Fatal("redirect followed")
	}
}
