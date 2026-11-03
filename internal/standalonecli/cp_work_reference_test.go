package standalonecli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/userapi"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func cpReferenceFixture(t *testing.T, kind string, hook func(*Service, *testStore, int), denied bool) (*Service, *atomic.Int32) {
	t.Helper()
	store := &testStore{present: true, credential: userauth.Credential{SessionID: "key-session", AccessToken: "key-access", RefreshToken: "key-refresh", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour)}}
	var calls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/graphql" {
			fmt.Fprint(w, `{"data":{"resolveCLIWorkReference":{"schemaVersion":"riido-work-reference.v1","workspaceId":"workspace-one","teamId":"team-one","id":"task-one","key":"SC-1","keyNumber":"1","componentType":"task"}}}`)
			return
		}
		if r.Method == http.MethodPost {
			if denied {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			purpose := "task_threads_read"
			if kind == "agent" {
				purpose = "task_agents_read"
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"schema_version": "riido-cli-" + kind + "-read.v1", "access_token": "riido_cli_" + kind + "_v1_" + strings.Repeat("A", 43), "token_type": "Bearer", "resource": "https://staging.ai-api.riido.io", "purpose": purpose, "expires_at": time.Now().Add(time.Minute).UTC().Format(time.RFC3339Nano)})
			return
		}
		if kind == "thread" {
			fmt.Fprint(w, `{"schema_version":"riido-ai-server.v1","task_id":"task-one","assigned_agent_ids":[],"threads":[]}`)
		} else {
			fmt.Fprint(w, `{"agents":[],"total":0,"cursor":"0","fetched_at":"2026-11-01T01:00:00Z","policy_checked":false}`)
		}
	}))
	t.Cleanup(server.Close)
	var service *Service
	client := &http.Client{Transport: protectedRoundTripper(func(r *http.Request) (*http.Response, error) {
		want := "monolith.api.riido.io"
		if r.Method == http.MethodGet {
			want = "staging.ai-api.riido.io"
		}
		if r.URL.Host != want || r.URL.Scheme != "https" {
			t.Fatal("unexpected origin")
		}
		deadline, ok := r.Context().Deadline()
		if !ok || time.Until(deadline) > 30*time.Second {
			t.Error("hop escaped compound timeout")
		}
		out := r.Clone(r.Context())
		u := *r.URL
		target, _ := url.Parse(server.URL)
		u.Scheme, u.Host = target.Scheme, target.Host
		out.URL, out.Host = &u, target.Host
		res, err := server.Client().Transport.RoundTrip(out)
		if err == nil && hook != nil {
			n := int(calls.Load())
			res.Body = &compoundResponseBody{ReadCloser: res.Body, afterClose: func() { hook(service, store, n) }}
		}
		return res, err
	})}
	var err error
	service, err = New(Config{APIOrigin: "https://monolith.api.riido.io", Store: store, HTTPClient: client})
	if err != nil {
		t.Fatal(err)
	}
	return service, &calls
}

func cpReferenceRead(ctx context.Context, service *Service, kind, work string) (bool, error) {
	if kind == "thread" {
		result, err := service.Threads(ctx, "workspace-one", work)
		return result != nil, err
	}
	result, err := service.Agents(ctx, "workspace-one", work, "")
	return result != nil, err
}

func TestCPWorkReferenceLoginAndCurrentAuthority(t *testing.T) {
	for _, kind := range []string{"thread", "agent"} {
		t.Run(kind, func(t *testing.T) {
			service, calls := cpReferenceFixture(t, kind, func(_ *Service, s *testStore, _ int) {
				if s.mu.TryLock() {
					s.mu.Unlock()
					t.Error("lock released before parsing")
				}
			}, false)
			for _, key := range []bool{true, false} {
				work, want := "SC-1", int32(3)
				if !key {
					work, want = "task-one", 2
				}
				before := calls.Load()
				present, err := cpReferenceRead(t.Context(), service, kind, work)
				if err != nil || !present || calls.Load()-before != want {
					t.Fatal("key/ID compound read failed", err, calls.Load()-before)
				}
			}
			service, calls = cpReferenceFixture(t, kind, nil, true)
			present, err := cpReferenceRead(t.Context(), service, kind, "SC-1")
			if present || !errors.Is(err, userapi.ErrPermission) || calls.Load() != 2 {
				t.Fatal("resolver bypassed current permission", present, err, calls.Load())
			}
		})
	}
}

func TestCPWorkReferenceDiscardsReplacedLoginAndCancellation(t *testing.T) {
	for _, kind := range []string{"thread", "agent"} {
		for _, after := range []int{1, 2, 3} {
			for _, cancelled := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/after%d/cancel%t", kind, after, cancelled), func(t *testing.T) {
					ctx, cancel := context.WithCancel(t.Context())
					defer cancel()
					service, calls := cpReferenceFixture(t, kind, func(_ *Service, s *testStore, n int) {
						if n == after {
							if cancelled {
								cancel()
							} else {
								s.credential.SessionID = "replacement"
								s.credential.AccessToken = "replacement-access"
							}
						}
					}, false)
					present, err := cpReferenceRead(ctx, service, kind, "SC-1")
					want := userauth.ErrLoggedOut
					if cancelled {
						want = context.Canceled
					}
					if present || !errors.Is(err, want) || calls.Load() != int32(after) {
						t.Fatal("mixed login result or extra call", present, err, calls.Load())
					}
					if _, err := service.manager.AccessToken(t.Context()); err != nil {
						t.Fatal("store lock retained", err)
					}
				})
			}
		}
	}
}

func TestCPWorkReferenceRejectsBeforeCredentials(t *testing.T) {
	for _, origin := range []string{"https://unsupported.invalid", "https://monolith.api.riido.io"} {
		store := &cpPreflightStore{testStore: &testStore{}}
		var calls atomic.Int32
		service, err := New(Config{APIOrigin: origin, Store: store, HTTPClient: &http.Client{Transport: protectedRoundTripper(func(*http.Request) (*http.Response, error) {
			calls.Add(1)
			return nil, io.EOF
		})}})
		if err != nil {
			t.Fatal(err)
		}
		cases := []func() error{
			func() error { _, e := service.Threads(t.Context(), "", "SC-1"); return e },
			func() error { _, e := service.Threads(t.Context(), "workspace-one", "task-one", "team-one"); return e },
			func() error { _, e := service.Thread(t.Context(), "workspace-one", "SC-1", "bad/id"); return e },
			func() error { _, e := service.Agents(t.Context(), "workspace-one", "SC-1", "01"); return e },
			func() error { _, e := service.Agent(t.Context(), "workspace-one", "SC-1", "bad/id"); return e },
		}
		if origin == "https://unsupported.invalid" {
			cases = append(cases, func() error { _, e := service.Threads(t.Context(), "workspace-one", "SC-1"); return e }, func() error { _, e := service.Agents(t.Context(), "workspace-one", "SC-1", ""); return e })
		}
		for _, run := range cases {
			if err := run(); err == nil || errors.Is(err, userauth.ErrNotAuthenticated) || calls.Load() != 0 || store.loads != 0 || store.locks != 0 {
				t.Fatal("invalid request reached credentials/network", err, calls.Load(), store.loads)
			}
		}
	}
}

type cpPreflightStore struct {
	*testStore
	loads, locks int
}

func (s *cpPreflightStore) Load(ctx context.Context) (userauth.Credential, error) {
	s.loads++
	return s.testStore.Load(ctx)
}

func (s *cpPreflightStore) Lock(ctx context.Context) (func(), error) {
	s.locks++
	return s.testStore.Lock(ctx)
}
