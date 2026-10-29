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

const userMeetingScopeFixture = `{"id":"meeting-one","teamId":"team-one","status":"available","details":{"id":"meeting-one","title":"미팅","meetingParticipants":[{"id":"row-one","memberOrder":0,"teamMember":null}],"meetingAttendees":[],"isAllDay":false},"labels":[{"id":"label-one","labelName":"회고","labelOrder":0}],"board":{"version":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","offset":0,"nextCursor":null,"complete":true,"groups":[{"id":"junction-one","meetingBoardGroupOrder":0,"meetingBoardGroup":{"id":"group-one","groupName":"논의","groupOrder":0}}],"cards":[{"id":"card-one","groupId":"junction-one","body":"보드 본문","boardOrder":0,"teamMember":null,"component":null,"conversedDate":null}]},"body":{"resourceId":"meeting-one","part":"document","status":"available","source":"provider_html","format":"html","version":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","offset":0,"content":"<p>미팅 에디터</p>","nextOffset":null,"complete":true},"descriptionBody":{"resourceId":"meeting-one","part":"description","status":"empty","source":"meeting_description","format":"html","version":"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","offset":0,"content":"","nextOffset":null,"complete":true},"partsStatus":[{"name":"board","status":"available","complete":true}]}`

func TestUserMeetingReadsStructuredBoardAndEmptyParticipantRows(t *testing.T) {
	calls := 0
	service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		var req struct {
			Query     string
			Variables map[string]any
		}
		if json.NewDecoder(r.Body).Decode(&req) != nil || r.Header.Get("Authorization") != "Bearer fixture-access" || req.Variables["meetingId"] != "meeting-one" || req.Variables["teamId"] != "team-one" || req.Variables["boardLimit"] != float64(100) {
			t.Fatal("meeting request lost authentication or scope")
		}
		for _, field := range []string{"getMeetingReadScope", "meetingParticipants", "meetingBoardGroup", "cards", "labels", "body", "descriptionBody"} {
			if !strings.Contains(req.Query, field) {
				t.Errorf("meeting query omitted %s", field)
			}
		}
		fmt.Fprintf(w, `{"data":{"getMeetingReadScope":%s}}`, userMeetingScopeFixture)
	})
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	if err := runMeeting(h.io, []string{"get", "meeting-one", "--team-id", "team-one", "--json"}); err != nil {
		t.Fatal(err)
	}
	data := decodeEnvelope(t, h.stdout.Bytes()).Data.(map[string]any)
	meeting := data["meeting"].(map[string]any)
	board := meeting["board"].(map[string]any)
	rows := meeting["details"].(map[string]any)["meetingParticipants"].([]any)
	if rows[0].(map[string]any)["teamMember"] != nil || len(board["groups"].([]any)) != 1 || board["cards"].([]any)[0].(map[string]any)["body"] != "보드 본문" || len(meeting["labels"].([]any)) != 1 || calls != 1 || len(h.daemon.requests) != 0 {
		t.Fatal("structured board or mode boundary was lost")
	}
}
