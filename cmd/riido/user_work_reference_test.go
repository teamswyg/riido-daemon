package main

import (
	"encoding/json"
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

func TestUserWorkKeyReadsResolvedTarget(t *testing.T) {
	store := &cliMemoryStore{present: true, credential: userauth.Credential{SessionID: "key-session", AccessToken: "key-access", RefreshToken: "key-refresh", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour)}}
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var q struct {
			Query     string
			Variables map[string]string
		}
		if r.URL.Path != "/graphql" || r.Header.Get("Authorization") != "Bearer key-access" || json.NewDecoder(r.Body).Decode(&q) != nil {
			t.Error("wrong read boundary")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if strings.Contains(q.Query, "resolveCLIWorkReference") {
			if q.Variables["reference"] != "SC-1" || q.Variables["teamId"] != "team-one" {
				t.Error("wrong reference scope", q.Variables)
			}
			fmt.Fprint(w, `{"data":{"resolveCLIWorkReference":{"schemaVersion":"riido-work-reference.v1","workspaceId":"workspace-one","teamId":"team-one","id":"work-one","key":"SC-1","keyNumber":"1","componentType":"task"}}}`)
			return
		}
		if q.Variables["componentId"] != "work-one" || q.Variables["teamId"] != "team-one" {
			t.Error("resolved target not used", q.Variables)
		}
		fmt.Fprint(w, `{"data":{"getComponent":{"id":"work-one","title":"Resolved work","componentMembers":[],"componentLabels":[]}}}`)
	}))
	t.Cleanup(server.Close)
	service, err := standalonecli.New(standalonecli.Config{APIOrigin: server.URL, Store: store, HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	err = runWork(h.io, []string{"get", "sc-1", "--team-id", "team-one", "--json"})
	if err != nil || calls != 2 || !strings.Contains(h.stdout.String(), "Resolved work") {
		t.Fatalf("key read failed: err=%v calls=%d output=%s", err, calls, h.stdout.String())
	}
	if len(h.daemon.requests) != 0 {
		t.Fatal("user key used agent daemon")
	}
}
