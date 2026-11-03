package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func TestDoctorExplicitUserAPIUsesWhoamiWithoutDaemonRequirement(t *testing.T) {
	store := &cliMemoryStore{present: true, credential: userauth.Credential{SessionID: "doctor-fixture-session", AccessToken: "fake-access-secret", RefreshToken: "fake-refresh-secret", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour)}}
	service, calls, opened := testUserService(t, store)
	h := newContextHarness(t, nil)
	h.daemon.down = true
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	if err := runDoctor(h.io, []string{"--check-api", "--json"}); err != nil {
		t.Fatal(err)
	}
	data := decodeEnvelope(t, h.stdout.Bytes()).Data.(map[string]any)
	if data["api_verified"] != true || *calls != 1 || *opened != 0 || store.confirmCalls != 0 {
		t.Fatalf("doctor did not verify API: data=%v calls=%d browser=%d", data, *calls, *opened)
	}
	if strings.Contains(h.stdout.String(), "fake-access-secret") || strings.Contains(h.stdout.String(), "user-one") {
		t.Fatal("diagnostics leaked credential or identity")
	}
}

func TestDoctorLocalAndAgentNeverConstructUserSession(t *testing.T) {
	for _, vars := range []map[string]string{nil, agentVars()} {
		h := newContextHarness(t, vars)
		h.daemon.down = true
		h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
			t.Fatal("diagnosis opened user storage")
			return nil, nil
		}
		_ = runDoctor(h.io, []string{"--json"})
		if vars != nil {
			h.stdout.Reset()
			_ = runDoctor(h.io, []string{"--check-api", "--json"})
		}
	}
}

func TestDoctorNoInputPropagatesUserPresenceDenial(t *testing.T) {
	h := newContextHarness(t, nil)
	h.io.userSession = func(args contextcli.Args) (*standalonecli.Service, error) {
		if !args.NoInput {
			t.Fatal("no-input was lost")
		}
		return nil, userauth.ErrStoreLocked
	}
	if got := exitCodeOf(t, runDoctor(h.io, []string{"--check-api", "--no-input", "--json"})); got != 3 {
		t.Fatalf("want user presence exit 3, got %d", got)
	}
	data := decodeEnvelope(t, h.stdout.Bytes()).Data.(map[string]any)
	if data["api_verified"] != false || data["healthy"] != false {
		t.Fatal("denied presence claimed success")
	}
}

func TestDoctorUserGraphQLErrorsPreserveExitWithoutLeakingBody(t *testing.T) {
	for code, want := range map[string]int{"authentication_required": 3, "permission_denied": 4, "not_found": 5, "unknown": 6} {
		t.Run(code, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				fmt.Fprintf(w, `{"errors":[{"message":"private-provider-detail","extensions":{"code":%q}}]}`, code)
			}))
			defer server.Close()
			store := &cliMemoryStore{present: true, credential: userauth.Credential{SessionID: "fixture-session", AccessToken: "fixture-access", RefreshToken: "fixture-refresh", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour)}}
			service, err := standalonecli.New(standalonecli.Config{APIOrigin: server.URL, Store: store})
			if err != nil {
				t.Fatal(err)
			}
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			if got := exitCodeOf(t, runDoctor(h.io, []string{"--check-api", "--json"})); got != want || calls != 1 {
				t.Fatalf("exit=%d calls=%d", got, calls)
			}
			if strings.Contains(h.stdout.String()+h.stderr.String(), "private-provider-detail") {
				t.Fatal("raw server error leaked")
			}
		})
	}
}

func TestDoctorLocalReportsInvalidOriginWithoutEchoingCredentials(t *testing.T) {
	h := newContextHarness(t, map[string]string{cliAPIOriginEnv: "https://private-user:private-secret@example.invalid/private?key=value"})
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
		t.Fatal("invalid origin accessed storage")
		return nil, errors.New("unreachable")
	}
	err := runDoctor(h.io, []string{"--json"})
	if err == nil {
		t.Fatal("invalid API origin reported healthy")
	}
	if strings.Contains(h.stdout.String()+h.stderr.String(), "private-secret") {
		t.Fatal("origin credentials leaked")
	}
	if doctorChecks(t, decodeEnvelope(t, h.stdout.Bytes()))["api_origin"] != "fail" {
		t.Fatal("missing origin failure")
	}
}

func TestDoctorUserAPIRefreshAndHTTPFailures(t *testing.T) {
	for _, scenario := range []string{"refresh", "revoked", "forbidden", "missing-identity"} {
		t.Run(scenario, func(t *testing.T) {
			calls, refreshes := 0, 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/oauth/token":
					refreshes++
					if err := r.ParseForm(); err != nil {
						t.Error(err)
					}
					if r.Form.Get("grant_type") != "urn:riido:params:oauth:grant-type:refresh-recovery" || len(r.Form.Get("recovery_secret")) != 43 {
						t.Error("wrong token grant")
					}
					if scenario == "revoked" {
						w.WriteHeader(http.StatusBadRequest)
						fmt.Fprint(w, `{"error":"invalid_grant","error_description":"private-revocation-reason"}`)
						return
					}
					fmt.Fprint(w, `{"access_token":"new-fixture-access","refresh_token":"new-fixture-refresh","token_type":"Bearer","expires_in":3600,"scope":"read","refresh_recovery":"v1"}`)
				case "/graphql":
					calls++
					if scenario == "forbidden" {
						w.WriteHeader(http.StatusForbidden)
						return
					}
					if scenario == "missing-identity" {
						fmt.Fprint(w, `{"data":{"getMyAccountInfo":null}}`)
						return
					}
					if r.Header.Get("Authorization") != "Bearer new-fixture-access" {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					fmt.Fprint(w, `{"data":{"getMyAccountInfo":{"id":"private-fixture-id"}}}`)
				default:
					t.Errorf("unexpected API endpoint: %s", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()
			store := &cliMemoryStore{present: true, credential: userauth.Credential{SessionID: "fixture-session", AccessToken: "fixture-access", RefreshToken: "fixture-refresh", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour)}}
			service, err := standalonecli.New(standalonecli.Config{APIOrigin: server.URL, Store: store})
			if err != nil {
				t.Fatal(err)
			}
			h := newContextHarness(t, map[string]string{cliAPIOriginEnv: server.URL})
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			err = runDoctor(h.io, []string{"--check-api", "--json"})
			data := decodeEnvelope(t, h.stdout.Bytes()).Data.(map[string]any)
			if scenario == "refresh" {
				if err != nil || calls != 2 || refreshes != 1 || data["api_verified"] != true {
					t.Fatalf("refresh failed: err=%v calls=%d refresh=%d", err, calls, refreshes)
				}
			} else {
				want := 6
				if scenario == "revoked" {
					want = 3
				}
				if scenario == "forbidden" {
					want = 4
				}
				if got := exitCodeOf(t, err); got != want || data["api_verified"] != false {
					t.Fatalf("exit=%d data=%v", got, data)
				}
				if calls != 1 {
					t.Fatalf("unexpected calls=%d", calls)
				}
				if scenario == "revoked" && (refreshes != 1 || store.present) {
					t.Fatal("revoked session not invalidated")
				}
			}
			for _, secret := range []string{"private-revocation-reason", "private-fixture-id", "new-fixture-access", "new-fixture-refresh"} {
				if strings.Contains(h.stdout.String()+h.stderr.String(), secret) {
					t.Fatal("diagnostic disclosed private response")
				}
			}
			if data["api_origin"] != server.URL || data["api_origin_source"] != "environment" {
				t.Fatal("selected origin missing")
			}
		})
	}
}

func TestDoctorHelpExposesExplicitUserAPICheck(t *testing.T) {
	var rootHelp bytes.Buffer
	printContextUsage(&rootHelp)
	h := newContextHarness(t, nil)
	if err := runDoctor(h.io, []string{"--help"}); err != nil {
		t.Fatal(err)
	}
	for _, help := range []string{rootHelp.String(), h.stderr.String()} {
		if !strings.Contains(help, "--check-api verifies user Whoami") || !strings.Contains(help, "agent API check is unsupported") {
			t.Fatal("doctor help hides the user API path")
		}
	}
}
