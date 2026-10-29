package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func TestUserContextIncludesAuthorizedMeetingSnapshotWithoutFollowingItsRelations(t *testing.T) {
	calls := []string{}
	service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Query     string
			Variables map[string]any
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(req.Query, "getComponentReadScope(") {
			calls = append(calls, "root")
			fixture := strings.Replace(userReadScopeFixture, `"meetings":[]`, `"meetings":[{"id":"meeting-one","teamId":"team-one","reasons":[{"kind":"related","relation":null}]}]`, 1)
			fmt.Fprintf(w, `{"data":{"getComponentReadScope":%s}}`, fixture)
			return
		}
		if !strings.Contains(req.Query, "getMeetingReadScope(") || req.Variables["meetingId"] != "meeting-one" || req.Variables["teamId"] != "team-one" || req.Variables["boardLimit"] != float64(100) {
			t.Fatalf("unexpected additional read: %v", req.Variables)
		}
		calls = append(calls, "meeting")
		fmt.Fprintf(w, `{"data":{"getMeetingReadScope":%s}}`, userMeetingScopeFixture)
	})
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	if err := runContext(h.io, []string{"get", "--work", "work-one", "--team-id", "team-one", "--json"}); err != nil {
		t.Fatal(err)
	}
	data := decodeEnvelope(t, h.stdout.Bytes()).Data.(map[string]any)
	meetings := data["context"].(map[string]any)["meetings"].([]any)
	if meetings[0].(map[string]any)["snapshot"] == nil || strings.Join(calls, ",") != "root,meeting" {
		t.Fatalf("selected meeting was not read: calls=%v", calls)
	}
	if len(h.daemon.requests) != 0 {
		t.Fatal("user read entered daemon broker")
	}
}

func TestUserContextRejectsMeetingWhosePermissionChangedAfterSelection(t *testing.T) {
	calls := 0
	service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			fixture := strings.Replace(userReadScopeFixture, `"meetings":[]`, `"meetings":[{"id":"meeting-one","teamId":"team-one","reasons":[{"kind":"related"}]}]`, 1)
			fmt.Fprintf(w, `{"data":{"getComponentReadScope":%s}}`, fixture)
			return
		}
		fmt.Fprint(w, `{"errors":[{"message":"private meeting title","extensions":{"code":"permission_denied"}}]}`)
	})
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	if got := exitCodeOf(t, runContext(h.io, []string{"get", "--work", "work-one", "--team-id", "team-one", "--json"})); got != 4 {
		t.Fatalf("exit=%d", got)
	}
	if calls != 2 || strings.Contains(h.stdout.String(), "private meeting title") || strings.Contains(h.stdout.String(), `"context"`) {
		t.Fatal("denied meeting returned partial context or private error")
	}
}
