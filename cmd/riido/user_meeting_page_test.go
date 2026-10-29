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

func TestUserMeetingFollowsOpaqueBoardCursorToCompletion(t *testing.T) {
	calls := 0
	service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		var req struct {
			Query     string
			Variables map[string]any
		}
		if json.NewDecoder(r.Body).Decode(&req) != nil || req.Variables["boardLimit"] != float64(1) || req.Variables["meetingId"] != "meeting-one" || req.Variables["teamId"] != "team-one" {
			t.Fatal("meeting page scope changed")
		}
		fixture := userMeetingScopeFixture
		if calls == 1 {
			if _, sent := req.Variables["boardCursor"]; sent {
				t.Fatal("invented first cursor")
			}
			fixture = strings.Replace(fixture, `"offset":0,"nextCursor":null,"complete":true`, `"offset":0,"nextCursor":"opaque-board-page-2","complete":false`, 1)
		} else {
			if req.Variables["boardCursor"] != "opaque-board-page-2" {
				t.Fatal("opaque cursor was replaced")
			}
			fixture = strings.Replace(fixture, `"offset":0,"nextCursor":null,"complete":true`, `"offset":1,"nextCursor":null,"complete":true`, 1)
			fixture = strings.Replace(fixture, `"id":"card-one"`, `"id":"card-two"`, 1)
		}
		fmt.Fprintf(w, `{"data":{"getMeetingReadScope":%s}}`, fixture)
	})
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	args := []string{"get", "meeting-one", "--team-id", "team-one", "--limit", "1", "--json"}
	if err := runMeeting(h.io, args); err != nil {
		t.Fatal(err)
	}
	first := decodeEnvelope(t, h.stdout.Bytes())
	m := first.Data.(map[string]any)["meeting"].(map[string]any)
	b := m["board"].(map[string]any)
	if m["status"] != "partial" || b["complete"] != false || b["nextCursor"] != "opaque-board-page-2" || len(first.Warnings) == 0 {
		t.Fatal("partial board was presented as complete")
	}
	if b["cards"].([]any)[0].(map[string]any)["groupId"] != b["groups"].([]any)[0].(map[string]any)["id"] {
		t.Fatal("card does not reference its group attachment")
	}
	h.stdout.Reset()
	if err := runMeeting(h.io, append(args, "--cursor", b["nextCursor"].(string))); err != nil {
		t.Fatal(err)
	}
	last := decodeEnvelope(t, h.stdout.Bytes())
	m = last.Data.(map[string]any)["meeting"].(map[string]any)
	b = m["board"].(map[string]any)
	if m["status"] != "available" || b["offset"] != float64(1) || b["complete"] != true || b["nextCursor"] != nil || len(last.Warnings) != 0 || calls != 2 {
		t.Fatal("last board page did not finish")
	}
	if b["cards"].([]any)[0].(map[string]any)["id"] != "card-two" || len(h.daemon.requests) != 0 {
		t.Fatal("page repeated or entered daemon")
	}
}
