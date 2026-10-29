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

const userReadScopeFixture = `{"schemaVersion":"riido-read-scope.v1","rootId":"work-one","teamId":"team-one","status":"available","selection":{"relatedWorkLimit":5,"relatedMeetingLimit":5,"childrenLimit":10},"works":[{"id":"work-one","kind":"component","teamId":"team-one","reasons":[{"kind":"root","relation":null}],"details":{"id":"work-one","title":"작업","estimate":0,"startDate":null,"isDone":false,"componentLabels":[]},"body":{"resourceId":"work-one","part":"document","status":"empty","source":"provider_json","format":"json","version":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","offset":0,"content":"","nextOffset":null,"complete":true},"descriptionBody":null,"partsStatus":[{"name":"body","status":"empty","complete":true}],"comments":{"resourceId":"work-one","resourceKind":"component","parentCommentId":null,"limit":20,"version":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","status":"available","page":1,"nextPage":null,"complete":true,"items":[]},"links":[],"github":[{"id":"pr-one","title":"연결 PR","number":7,"state":"open","commentsCount":0,"aiShortSummary":""}],"repositories":[],"sprints":[],"origin":null,"priorities":[],"checklists":[],"attachments":[],"figma":{"id":"figma-one","url":"https://www.figma.com/design/fixture","fileKey":"fixture","nodeId":null}}],"meetings":[]}`

func TestUserContextCallsFixedScopeAPIAndPreservesProductFields(t *testing.T) {
	calls := 0
	service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		var req struct {
			Query     string
			Variables map[string]string
		}
		if json.NewDecoder(r.Body).Decode(&req) != nil || r.Header.Get("Authorization") != "Bearer fixture-access" || req.Variables["componentId"] != "work-one" || req.Variables["teamId"] != "team-one" {
			t.Error("expected an authenticated, explicitly scoped request")
		}
		for _, field := range []string{"getComponentReadScope", "componentLabels", "github", "figma", "comments", "descriptionBody"} {
			if !strings.Contains(req.Query, field) {
				t.Errorf("fixed query omitted %s", field)
			}
		}
		fmt.Fprintf(w, `{"data":{"getComponentReadScope":%s}}`, userReadScopeFixture)
	})
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	if err := runContext(h.io, []string{"get", "--work", "work-one", "--team-id", "team-one", "--json"}); err != nil {
		t.Fatal(err)
	}
	env := decodeEnvelope(t, h.stdout.Bytes())
	data := env.Data.(map[string]any)
	scope := data["context"].(map[string]any)
	work := scope["works"].([]any)[0].(map[string]any)
	detail := work["details"].(map[string]any)
	if data["mode"] != "user" || data["live"] != true || detail["estimate"] != float64(0) || detail["isDone"] != false || detail["startDate"] != nil || work["figma"] == nil || len(work["github"].([]any)) != 1 || calls != 1 || len(h.daemon.requests) != 0 {
		t.Fatal("lost product values, mode boundary, or read count")
	}
}
