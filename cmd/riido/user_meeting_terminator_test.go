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

func TestUserMeetingDashIDAfterTerminator(t *testing.T) {
	const id = "-meeting-fixture"
	calls := 0
	service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		var req struct {
			Variables map[string]any `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Variables["meetingId"] != id || req.Variables["teamId"] != "team-one" {
			t.Fatal("opaque ID was not forwarded intact", err)
		}
		fmt.Fprintf(w, `{"data":{"getMeetingReadScope":%s}}`, strings.ReplaceAll(userMeetingScopeFixture, "meeting-one", id))
	})
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	if err := runMeeting(h.io, []string{"get", "--team-id", "team-one", "--json", "--", id}); err != nil {
		t.Fatal(err)
	}
	data := decodeEnvelope(t, h.stdout.Bytes()).Data.(map[string]any)
	if data["meeting"].(map[string]any)["id"] != id || calls != 1 || len(h.daemon.requests) != 0 {
		t.Fatal("meeting response or user/agent boundary changed")
	}
}

func TestWantsJSONRespectsOptionTerminator(t *testing.T) {
	if wantsJSON([]string{"--", "--json"}) || wantsJSON([]string{"--", "--format=json"}) || !wantsJSON([]string{"--json", "--", "-meeting-fixture"}) {
		t.Fatal("literal arguments changed output format")
	}
}
