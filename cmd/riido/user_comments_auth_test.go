package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

type commentGuardStore struct {
	*cliMemoryStore
	loadErr error
}

func (s *commentGuardStore) Load(ctx context.Context) (userauth.Credential, error) {
	if s.loadErr != nil {
		return userauth.Credential{}, s.loadErr
	}
	return s.cliMemoryStore.Load(ctx)
}

func TestUserCommentsProtectedStoreFailuresPrecedeAPI(t *testing.T) {
	for _, command := range []string{"comments", "replies"} {
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
			t.Run(command+"/"+tc.name, func(t *testing.T) {
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
					if tc.name == "no-input" && !args.NoInput {
						t.Fatal("no-input was lost before opening storage")
					}
					return service, nil
				}
				args := commentCommandArgs(command)
				if tc.name == "no-input" {
					args = append(args, "--no-input")
				}
				if got := exitCodeOf(t, runWork(h.io, args)); got != tc.exit || calls != 0 {
					t.Fatalf("exit=%d calls=%d", got, calls)
				}
			})
		}
	}
}

func TestUserCommentRevokedRefreshRequiresReauthentication(t *testing.T) {
	for _, command := range []string{"comments", "replies"} {
		t.Run(command, func(t *testing.T) {
			reads, refreshes := 0, 0
			service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/oauth/token" {
					refreshes++
					w.WriteHeader(http.StatusBadRequest)
					fmt.Fprint(w, `{"error":"invalid_grant","error_description":"private-grant-detail"}`)
					return
				}
				reads++
				w.WriteHeader(http.StatusUnauthorized)
			})
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			if got := exitCodeOf(t, runWork(h.io, commentCommandArgs(command))); got != 3 || reads != 1 || refreshes != 1 {
				t.Fatalf("exit=%d reads=%d refreshes=%d", got, reads, refreshes)
			}
		})
	}
}

func TestUserCommentReadsAfterLogoutCannotReuseStoredGrant(t *testing.T) {
	for _, command := range []string{"comments", "replies"} {
		t.Run(command, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.URL.Path != "/oauth/revoke" {
					t.Error("logged-out read reached API")
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()
			store := &cliMemoryStore{present: true, credential: userauth.Credential{SessionID: "session-one", AccessToken: "fixture-access", RefreshToken: "fixture-refresh", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour)}}
			service, err := standalonecli.New(standalonecli.Config{APIOrigin: server.URL, Store: store})
			if err != nil {
				t.Fatal(err)
			}
			if err := service.Logout(context.Background()); err != nil {
				t.Fatal(err)
			}
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			if got := exitCodeOf(t, runWork(h.io, commentCommandArgs(command))); got != 3 || calls != 1 || store.present {
				t.Fatalf("exit=%d calls=%d present=%v", got, calls, store.present)
			}
		})
	}
}
