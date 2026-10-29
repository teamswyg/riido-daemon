package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func TestUserRelationsReadsExistingLiveContract(t *testing.T) {
	calls := 0
	service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		var request struct {
			Query     string            `json:"query"`
			Variables map[string]string `json:"variables"`
		}
		if r.URL.Path != "/graphql" || r.Header.Get("Authorization") != "Bearer fixture-access" || json.NewDecoder(r.Body).Decode(&request) != nil || !strings.Contains(request.Query, "getComponentRelatedIssues") || !reflect.DeepEqual(request.Variables, map[string]string{"componentId": "work-one", "teamId": "team-one"}) {
			t.Error("expected one authenticated, scoped relation query")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fmt.Fprint(w, `{"data":{"getComponentRelatedIssues":[{"relationCategory":"blockedBy","issueId":"issue-other","issueType":"component","issue":{"id":"issue-other","component":{"id":"component-other","componentType":"task","title":"","isDone":false,"status":null},"backlog":null,"intake":null}}]}}`)
	})
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	if err := runWork(h.io, []string{"relations", "work-one", "--team-id", "team-one", "--json"}); err != nil {
		t.Fatal(err)
	}
	envelope := decodeEnvelope(t, h.stdout.Bytes())
	data := envelope.Data.(map[string]any)
	rows := data["relations"].([]any)
	row := rows[0].(map[string]any)
	issue := row["issue"].(map[string]any)
	component := issue["component"].(map[string]any)
	if calls != 1 || len(h.daemon.requests) != 0 || data["mode"] != "user" || envelope.Page != nil || row["relationCategory"] != "blockedBy" || row["issueId"] != issue["id"] || component["id"] != "component-other" || component["title"] != "" || component["isDone"] != false || component["status"] != nil {
		t.Fatalf("relation contract changed: %s", h.stdout.String())
	}
	if strings.Contains(h.stdout.String()+h.stderr.String(), "fixture-access") {
		t.Fatal("credential leaked")
	}
}
