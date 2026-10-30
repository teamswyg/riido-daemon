package userapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func searchPageFixture() map[string]any {
	return map[string]any{"schemaVersion": "riido-work-search.v1", "workspaceId": "workspace-one", "query": "검색", "limit": 2, "complete": false, "nextCursor": "bmV4dC1wYWdl", "items": []any{
		map[string]any{"kind": "component", "id": "work-one", "teamId": "team-one", "title": "검색 작업", "componentType": "task", "keyNumber": "1", "key": "RD-1", "updatedAt": "2026-10-31T01:02:03.123456Z", "matchedFields": []any{"title", "document"}},
		map[string]any{"kind": "intake", "id": "intake-one", "teamId": "team-one", "title": "검색 인테이크", "componentType": nil, "keyNumber": nil, "key": nil, "updatedAt": "2026-10-31T01:02:03.123456Z", "matchedFields": []any{"title"}},
	}}
}

func searchFixtureJSON(f map[string]any) string {
	raw, _ := json.Marshal(map[string]any{"data": map[string]any{"searchWorkspaceWork": f}})
	return string(raw)
}

func TestWorkSearchFixedReadAndCursor(t *testing.T) {
	calls := 0
	r := Reader{Endpoint: "https://api.example/graphql", Client: doerFunc(func(_ context.Context, req *http.Request) (*http.Response, error) {
		calls++
		var payload struct {
			Query     string
			Variables map[string]any
		}
		if req.Method != http.MethodPost || json.NewDecoder(req.Body).Decode(&payload) != nil {
			t.Fatal("invalid request")
		}
		if !strings.Contains(payload.Query, "query RiidoCLIWorkSearch") || strings.Contains(payload.Query, "mainSearch") || strings.Contains(payload.Query, "mutation") || payload.Variables["workspaceId"] != "workspace-one" || payload.Variables["query"] != "검색" || payload.Variables["cursor"] != "prior-page" || payload.Variables["limit"] != float64(2) {
			t.Fatal("changed search contract")
		}
		return response(http.StatusOK, searchFixtureJSON(searchPageFixture())), nil
	})}
	got, err := r.Search(t.Context(), WorkSearchRequest{WorkspaceID: "workspace-one", Query: "  검색  ", Cursor: "prior-page", Limit: 2})
	if err != nil || got.Complete || len(got.Items) != 2 || got.NextCursor == nil || got.Items[1].Key != nil || calls != 1 {
		t.Fatalf("bad search: %+v %v", got, err)
	}
}

func TestWorkSearchRejectsMalformedOrPartialResults(t *testing.T) {
	for name, change := range map[string]func(map[string]any){
		"shell-next":   func(f map[string]any) { f["nextCursor"] = "$(echo unsafe)" },
		"newline-next": func(f map[string]any) { f["nextCursor"] = "cGFnZS10d28\n" },
		"workspace":    func(f map[string]any) { f["workspaceId"] = "other" }, "query": func(f map[string]any) { f["query"] = "other" }, "schema": func(f map[string]any) { f["schemaVersion"] = "future" }, "limit": func(f map[string]any) { f["limit"] = 20 }, "missing-items": func(f map[string]any) { delete(f, "items") }, "missing-complete": func(f map[string]any) { delete(f, "complete") }, "missing-next": func(f map[string]any) { f["nextCursor"] = nil }, "terminal-next": func(f map[string]any) { f["complete"] = true }, "short-incomplete": func(f map[string]any) { f["items"] = f["items"].([]any)[:1] }, "duplicate": func(f map[string]any) { f["items"].([]any)[1] = f["items"].([]any)[0] }, "invalid-kind": func(f map[string]any) { f["items"].([]any)[0].(map[string]any)["kind"] = "meeting" }, "missing-title": func(f map[string]any) { delete(f["items"].([]any)[0].(map[string]any), "title") }, "unknown-match": func(f map[string]any) { f["items"].([]any)[0].(map[string]any)["matchedFields"] = []any{"private"} }, "reordered": func(f map[string]any) { items := f["items"].([]any); items[0], items[1] = items[1], items[0] },
	} {
		t.Run(name, func(t *testing.T) {
			f := searchPageFixture()
			change(f)
			got, err := relationReader(searchFixtureJSON(f)).Search(t.Context(), WorkSearchRequest{WorkspaceID: "workspace-one", Query: "검색", Limit: 2})
			if got != nil || !errors.Is(err, ErrInvalidResponse) {
				t.Fatalf("accepted %+v %v", got, err)
			}
		})
	}
}
