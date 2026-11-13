package standalonecli

import (
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

func workspaceLocatorService(t *testing.T, replaceAfterResolve, denyRead bool) (*Service, *atomic.Int32) {
	t.Helper()
	store := &testStore{present: true, credential: userauth.Credential{SessionID: "locator-session", AccessToken: "locator-access", RefreshToken: "locator-refresh", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour)}}
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		var q struct {
			Query     string
			Variables map[string]any
		}
		if r.URL.Path != "/graphql" || r.Header.Get("Authorization") != "Bearer locator-access" || json.NewDecoder(r.Body).Decode(&q) != nil || !strings.Contains(q.Query, "locatorVersion:2") {
			t.Error("changed fixed authorized request")
			w.WriteHeader(400)
			return
		}
		if store.mu.TryLock() {
			store.mu.Unlock()
			t.Error("session lock released")
		}
		if q.Variables["workspaceId"] != "workspace-one" {
			t.Error("workspace lost")
		}
		if strings.Contains(q.Query, "resolveCLIWorkReference") {
			if q.Variables["reference"] != "work-one" {
				t.Error("canonical ID skipped resolution")
			}
			fmt.Fprint(w, `{"data":{"resolveCLIWorkReference":{"schemaVersion":"riido-work-reference.v2","workspaceId":"workspace-one","teamId":null,"id":"work-one","key":"SC-1","keyNumber":"1","componentType":"project"}}}`)
			return
		}
		if q.Variables["componentId"] != "work-one" || q.Variables["teamId"] != nil {
			t.Error("locator changed")
		}
		if denyRead {
			fmt.Fprint(w, `{"data":{"getComponent":null},"errors":[{"extensions":{"code":"permission_denied"}}]}`)
			return
		}
		fmt.Fprint(w, `{"data":{"getComponent":{"schemaVersion":"riido-work-detail.v2","workspaceId":"workspace-one","ownerTeamId":null,"canEdit":true,"id":"work-one","componentType":"project","title":"workspace project","componentMembers":[],"componentLabels":[]}}}`)
	}))
	t.Cleanup(server.Close)
	client := server.Client()
	transport := client.Transport
	client.Transport = protectedRoundTripper(func(r *http.Request) (*http.Response, error) {
		res, err := transport.RoundTrip(r)
		if err == nil && replaceAfterResolve && calls.Load() == 1 {
			res.Body = &compoundResponseBody{ReadCloser: res.Body, afterClose: func() {
				store.credential.SessionID = "replacement"
				store.credential.AccessToken = "replacement-access"
			}}
		}
		return res, err
	})
	s, err := New(Config{APIOrigin: server.URL, Store: store, HTTPClient: client})
	if err != nil {
		t.Fatal(err)
	}
	return s, &calls
}

func TestWorkspaceLocatorV2ProtectedSession(t *testing.T) {
	for _, tc := range []struct {
		name          string
		replace, deny bool
		want          error
		calls         int32
	}{{"workspace-project", false, false, nil, 2}, {"permission-rechecked", false, true, userapi.ErrPermission, 2}, {"replaced-login", true, false, userauth.ErrLoggedOut, 1}} {
		t.Run(tc.name, func(t *testing.T) {
			s, calls := workspaceLocatorService(t, tc.replace, tc.deny)
			got, err := s.WorkV2(t.Context(), userapi.WorkReferenceRequest{Reference: "work-one", WorkspaceID: "workspace-one"})
			if !errors.Is(err, tc.want) || calls.Load() != tc.calls {
				t.Fatal(got, err, calls.Load())
			}
			if err != nil && got.ID != "" {
				t.Fatal("partial output escaped")
			}
			if err == nil && (got.ID != "work-one" || got.OwnerTeamID != nil) {
				t.Fatal("workspace locator lost")
			}
		})
	}
}
