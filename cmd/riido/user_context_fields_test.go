package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func TestUserContextPreservesCurrentInspectorFields(t *testing.T) {
	var scope map[string]any
	if err := json.Unmarshal([]byte(userReadScopeFixture), &scope); err != nil {
		t.Fatal(err)
	}
	work := scope["works"].([]any)[0].(map[string]any)
	work["sprints"] = []any{map[string]any{"id": "sprint-one", "number": 3, "title": "개발", "startDate": "2026-10-26", "endDate": "2026-10-30"}}
	work["origin"] = map[string]any{"id": "intake-one", "kind": "intake", "title": "원 요청", "isDeleted": false}
	work["links"] = []any{map[string]any{"id": "link-one", "url": "https://example.test", "title": "자료", "description": "저장된 설명", "createdAt": "2026-10-29T00:00:00Z", "logo": nil}}
	service := commentService(t, func(out http.ResponseWriter, r *http.Request) {
		var req struct{ Query string }
		if json.NewDecoder(r.Body).Decode(&req) != nil {
			t.Fatal("invalid request")
		}
		for _, field := range []string{"sprints{", "origin{", "description createdAt logo", "estimate estimateMethod"} {
			if !strings.Contains(req.Query, field) {
				t.Errorf("fixed query is missing %s", field)
			}
		}
		_ = json.NewEncoder(out).Encode(map[string]any{"data": map[string]any{"getComponentReadScope": scope}})
	})
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	if err := runContext(h.io, []string{"get", "--work", "work-one", "--team-id", "team-one", "--json"}); err != nil {
		t.Fatal(err)
	}
	got := decodeEnvelope(t, h.stdout.Bytes()).Data.(map[string]any)["context"].(map[string]any)["works"].([]any)[0].(map[string]any)
	if got["sprints"] == nil || got["origin"] == nil {
		t.Fatal("current inspector sprint/origin was dropped")
	}
	link := got["links"].([]any)[0].(map[string]any)
	if link["description"] != "저장된 설명" || link["logo"] != nil {
		t.Fatal("stored link detail was changed")
	}
}
