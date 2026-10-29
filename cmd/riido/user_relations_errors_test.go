package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func relationArgs() []string {
	return []string{"relations", "work-one", "--team-id", "team-one", "--json"}
}

func TestUserRelationsFailuresDoNotLeakPartialDataOrRetry(t *testing.T) {
	for _, tc := range []struct {
		name, code   string
		status, exit int
	}{
		{"forbidden", "", 403, 4},
		{"rate-limit", "", 429, 7},
		{"server", "", 500, 6},
		{"permission", "permission_denied", 200, 4},
		{"missing", "not_found", 200, 5},
		{"auth", "authentication_required", 200, 3},
		{"unknown", "private-code", 200, 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.URL.Path != "/graphql" {
					t.Error("failure attempted refresh")
				}
				w.Header().Set("Retry-After", "600")
				w.WriteHeader(tc.status)
				fmt.Fprintf(w, `{"data":{"getComponentRelatedIssues":[],"private":"private-body"},"errors":[{"message":"private-detail","extensions":{"code":%q}}]}`, tc.code)
			})
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			if got := exitCodeOf(t, runWork(h.io, relationArgs())); got != tc.exit {
				t.Fatalf("exit=%d want=%d", got, tc.exit)
			}
			envelope := decodeEnvelope(t, h.stdout.Bytes())
			if calls != 1 || envelope.Data != nil || envelope.Error == nil || strings.Contains(h.stdout.String()+h.stderr.String(), "private-") {
				t.Fatalf("unexpected partial result/retry: calls=%d", calls)
			}
			if tc.status == 429 && (envelope.Error.Code != contextcli.CodeRateLimited || !envelope.Error.Retryable) {
				t.Fatal("rate limit contract changed")
			}
		})
	}
}

func TestUserRelationsRefreshReplaysSameReadOnce(t *testing.T) {
	reads, refreshes := 0, 0
	var first string
	service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			refreshes++
			fmt.Fprint(w, `{"access_token":"refreshed-access","refresh_token":"refreshed-refresh","token_type":"Bearer","expires_in":3600,"scope":"read"}`)
			return
		}
		reads++
		body, _ := io.ReadAll(r.Body)
		if reads == 1 {
			first = string(body)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if first != string(body) || r.Header.Get("Authorization") != "Bearer refreshed-access" {
			t.Error("refresh changed read identity")
		}
		fmt.Fprint(w, `{"data":{"getComponentRelatedIssues":[]}}`)
	})
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	if err := runWork(h.io, relationArgs()); err != nil {
		t.Fatal(err)
	}
	if reads != 2 || refreshes != 1 || len(h.daemon.requests) != 0 {
		t.Fatalf("reads=%d refresh=%d", reads, refreshes)
	}
}

func TestUserRelationsProtectedStoreFailuresPrecedeAPI(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		exit int
	}{
		{"logged-out", userauth.ErrNotAuthenticated, 3},
		{"no-input", userauth.ErrStoreLocked, 3},
		{"helper-unavailable", securecredential.ErrUnavailable, 8},
		{"cancelled", context.Canceled, 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls++ }))
			defer server.Close()
			store := &commentGuardStore{cliMemoryStore: &cliMemoryStore{}, loadErr: tc.err}
			service, err := standalonecli.New(standalonecli.Config{APIOrigin: server.URL, Store: store})
			if err != nil {
				t.Fatal(err)
			}
			h := newContextHarness(t, nil)
			h.io.userSession = func(args contextcli.Args) (*standalonecli.Service, error) {
				if !args.NoInput {
					t.Fatal("no-input lost")
				}
				return service, nil
			}
			if got := exitCodeOf(t, runWork(h.io, append(relationArgs(), "--no-input"))); got != tc.exit || calls != 0 {
				t.Fatalf("exit=%d calls=%d", got, calls)
			}
		})
	}
}

func TestUserRelationsLogoutRemovesLocalReadCredential(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/oauth/revoke" {
			t.Error("logged-out read reached API")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	store := &cliMemoryStore{present: true, credential: userauth.Credential{SessionID: "fixture-session", AccessToken: "fixture-access", RefreshToken: "fixture-refresh", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour)}}
	service, err := standalonecli.New(standalonecli.Config{APIOrigin: server.URL, Store: store})
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Logout(context.Background()); err != nil {
		t.Fatal(err)
	}
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	if got := exitCodeOf(t, runWork(h.io, relationArgs())); got != 3 || calls != 1 || store.present {
		t.Fatalf("exit=%d calls=%d retained=%t", got, calls, store.present)
	}
}
