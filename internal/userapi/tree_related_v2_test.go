package userapi

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func treeV2Fixture(t *testing.T) map[string]any {
	t.Helper()
	var v map[string]any
	if err := json.Unmarshal([]byte(`{"schemaVersion":"riido-read-scope.v2","workspaceId":"workspace-one","rootId":"work-one","teamId":null,"status":"available","selection":{"childrenLimit":10},"works":[{"id":"work-one","workspaceId":"workspace-one","kind":"component","teamId":null,"reasons":[{"kind":"root"}],"summary":{"id":"work-one","kind":"component","title":"Project","keyNumber":null,"componentType":"project"}},{"id":"child-one","workspaceId":"workspace-one","kind":"intake","teamId":"team-one","reasons":[{"kind":"child"}],"summary":{"id":"child-one","kind":"intake","title":"Intake","keyNumber":"SC-2","componentType":"task"}}]}`), &v); err != nil {
		t.Fatal(err)
	}
	return v
}

func TestWorkspaceLocatorV2TreeAndBoundaries(t *testing.T) {
	locator := WorkLocator{ID: "work-one", WorkspaceID: "workspace-one", ComponentType: "project"}
	for _, tc := range []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"valid", nil},
		{"missing-team", func(v map[string]any) { delete(v, "teamId") }},
		{"wrong-workspace", func(v map[string]any) { v["workspaceId"] = "other" }},
		{"invented-team", func(v map[string]any) { v["teamId"] = "fake-team" }},
		{"duplicate-current-id", func(v map[string]any) { rows := v["works"].([]any); v["works"] = append(rows, rows[1]) }},
		{"child-without-team", func(v map[string]any) { v["works"].([]any)[1].(map[string]any)["teamId"] = nil }},
		{"child-cross-workspace", func(v map[string]any) { v["works"].([]any)[1].(map[string]any)["workspaceId"] = "other" }},
		{"wrong-version", func(v map[string]any) { v["schemaVersion"] = "riido-read-scope.v1" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := treeV2Fixture(t)
			if tc.mutate != nil {
				tc.mutate(v)
			}
			r := locatorV2Reader(t, "getComponentReadScope", v, func(q string, vars map[string]any) {
				if vars["teamId"] != nil || vars["workspaceId"] != locator.WorkspaceID || strings.Contains(q, "body{") {
					t.Fatal("incorrect/minimally scoped projection", q, vars)
				}
			})
			out, err := r.TreeV2(t.Context(), locator)
			if tc.mutate != nil {
				if !errors.Is(err, ErrInvalidResponse) {
					t.Fatal("accepted invalid tree", out, err)
				}
				return
			}
			if err != nil || out.Root.TeamID != nil || len(out.Children) != 1 || out.Children[0].Kind != "intake" || out.MaxDepth != 1 {
				t.Fatal(out, err)
			}
		})
	}
}

func TestWorkspaceLocatorV2RelatedReads(t *testing.T) {
	locator := WorkLocator{ID: "work-one", WorkspaceID: "workspace-one", ComponentType: "project"}
	check := func(q string, v map[string]any) {
		if v["teamId"] != nil || v["workspaceId"] != locator.WorkspaceID || !strings.Contains(q, "workspaceId:$workspaceId,locatorVersion:2") {
			t.Fatal("lost locator", q, v)
		}
	}
	r := locatorV2Reader(t, "getComponentRelatedIssues", []any{}, check)
	if rows, err := r.RelationsV2(t.Context(), locator); err != nil || rows == nil {
		t.Fatal(rows, err)
	}
	r = locatorV2Reader(t, "getComponentActivityLogs", map[string]any{"activityLogs": []any{}, "meta": map[string]any{"total": 0, "currentPage": 1, "nextPage": nil, "prevPage": nil, "lastPage": 0}}, check)
	if _, err := r.ActivityV2(t.Context(), locator, ActivityPagination{Page: "1", Limit: "20"}); err != nil {
		t.Fatal(err)
	}
}

func TestWorkspaceLocatorV2StageSearchKeys(t *testing.T) {
	for _, kind := range []string{"backlog", "intake"} {
		for _, version := range []int{1, 2} {
			t.Run(kind+locatorSchema("", version), func(t *testing.T) {
				p := map[string]any{"schemaVersion": locatorSchema("riido-work-search", version), "workspaceId": "workspace-one", "query": "sc-2", "limit": 20, "complete": true, "nextCursor": nil, "items": []any{map[string]any{"id": "work-one", "kind": kind, "teamId": "team-one", "componentType": "task", "title": "current task", "key": "SC-2", "keyNumber": "2", "matchedFields": []string{"key"}, "updatedAt": "2026-11-18T00:00:00Z"}}}
				r := locatorV2Reader(t, "searchWorkspaceWork", p, nil)
				// Native TASK keys are additive only in the explicitly opted-in v2 contract.
				if version == 1 {
					r = relationReader(searchFixtureJSON(p))
					_, err := r.Search(t.Context(), WorkSearchRequest{WorkspaceID: "workspace-one", Query: "sc-2", LocatorVersion: 1})
					if !errors.Is(err, ErrInvalidResponse) {
						t.Fatal("v1 accepted native stage fields", err)
					}
					return
				}
				out, err := r.Search(t.Context(), WorkSearchRequest{WorkspaceID: "workspace-one", Query: "sc-2", LocatorVersion: 2})
				if err != nil || *out.Items[0].Key != "SC-2" {
					t.Fatal(out, err)
				}
			})
		}
	}
}
