package standalonecli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/userapi"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func workKeyFixture(t *testing.T, hook func(*Service, *testStore, int), refuseRead, refresh bool) (*Service, *atomic.Int32) {
	t.Helper()
	store := &testStore{present: true, credential: userauth.Credential{SessionID: "key-session", AccessToken: "key-access", RefreshToken: "key-refresh", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour)}}
	var calls, refreshes atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			refreshes.Add(1)
			fmt.Fprint(w, `{"access_token":"next-access","refresh_token":"next-refresh","token_type":"Bearer","expires_in":3600,"refresh_recovery":"v1"}`)
			return
		}
		n := calls.Add(1)
		want := "Bearer key-access"
		if refreshes.Load() > 0 {
			want = "Bearer next-access"
		}
		if r.URL.Path != "/graphql" || r.Header.Get("Authorization") != want {
			t.Error("origin/login changed")
		}
		if refresh && n == 2 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var q struct {
			Query     string
			Variables map[string]string
		}
		if json.NewDecoder(r.Body).Decode(&q) != nil {
			t.Error("bad query")
			return
		}
		if strings.Contains(q.Query, "resolveCLIWorkReference") {
			if q.Variables["reference"] != "SC-1" || q.Variables["workspaceId"] != "workspace-one" {
				t.Error("lost reference scope", q.Variables)
			}
			fmt.Fprint(w, `{"data":{"resolveCLIWorkReference":{"schemaVersion":"riido-work-reference.v1","workspaceId":"workspace-one","teamId":"team-one","id":"work-one","key":"SC-1","keyNumber":"SC-1","componentType":"task"}}}`)
			return
		}
		if q.Variables["componentId"] != "work-one" || q.Variables["teamId"] != "team-one" {
			t.Error("unresolved read", q.Variables)
		}
		if refuseRead {
			fmt.Fprint(w, `{"data":{"getComponent":null},"errors":[{"message":"hidden-canary","extensions":{"code":"permission_denied"}}]}`)
			return
		}
		fmt.Fprint(w, `{"data":{"getComponent":{"id":"work-one","title":"key-work","componentMembers":[],"componentLabels":[]}}}`)
	}))
	t.Cleanup(server.Close)
	var service *Service
	client := server.Client()
	transport := client.Transport
	client.Transport = protectedRoundTripper(func(r *http.Request) (*http.Response, error) {
		res, err := transport.RoundTrip(r)
		if err == nil && r.URL.Path == "/graphql" && res.StatusCode == http.StatusOK && hook != nil {
			n := int(calls.Load())
			res.Body = &compoundResponseBody{ReadCloser: res.Body, afterClose: func() { hook(service, store, n) }}
		}
		return res, err
	})
	var err error
	service, err = New(Config{APIOrigin: server.URL, Store: store, HTTPClient: client})
	if err != nil {
		t.Fatal(err)
	}
	return service, &calls
}

func TestWorkKeySessionAndRevocation(t *testing.T) {
	for _, refresh := range []bool{false, true} {
		service, calls := workKeyFixture(t, func(_ *Service, s *testStore, _ int) {
			if s.mu.TryLock() {
				s.mu.Unlock()
				t.Error("login lock released during parsing")
			}
		}, false, refresh)
		work, err := service.Work(t.Context(), "SC-1", "", "workspace-one")
		want := int32(2)
		if refresh {
			want = 3
		}
		if err != nil || work.ID != "work-one" || calls.Load() != want {
			t.Fatal("key/read session failed", work, err, calls.Load())
		}
	}
	service, calls := workKeyFixture(t, nil, true, false)
	work, err := service.Work(t.Context(), "SC-1", "", "workspace-one")
	if work.ID != "" || !errors.Is(err, userapi.ErrPermission) || calls.Load() != 2 {
		t.Fatal("resolver authority replaced current access", work, err)
	}
}

func TestWorkKeyDiscardsReplacedLoginAndCancellation(t *testing.T) {
	for _, after := range []int{1, 2} {
		for _, cancelled := range []bool{false, true} {
			t.Run(fmt.Sprintf("after%d/cancel%t", after, cancelled), func(t *testing.T) {
				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()
				service, calls := workKeyFixture(t, func(_ *Service, s *testStore, n int) {
					if n == after {
						if cancelled {
							cancel()
						} else {
							s.credential.SessionID = "replacement"
							s.credential.AccessToken = "replacement-access"
						}
					}
				}, false, false)
				work, err := service.Work(ctx, "SC-1", "", "workspace-one")
				want := userauth.ErrLoggedOut
				if cancelled {
					want = context.Canceled
				}
				if work.ID != "" || !errors.Is(err, want) || calls.Load() != int32(after) {
					t.Fatal("stale result escaped", work, err, calls.Load())
				}
				if _, err := service.manager.AccessToken(t.Context()); err != nil {
					t.Fatal("retained read lock", err)
				}
			})
		}
	}
}

func TestWorkKeyIDPathHasNoLookup(t *testing.T) {
	service, calls := workKeyFixture(t, nil, false, false)
	work, err := service.Work(t.Context(), "work-one", "team-one", "unrelated-default")
	if err != nil || work.ID != "work-one" || calls.Load() != 1 {
		t.Fatal("ID path added lookup or workspace binding", work, err)
	}
}
