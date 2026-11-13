package userapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func locatorV2Reader(t *testing.T, field string, payload any, check func(string, map[string]any)) Reader {
	t.Helper()
	return Reader{Endpoint: "https://api.example/graphql", Client: doerFunc(func(_ context.Context, req *http.Request) (*http.Response, error) {
		var q struct {
			Query     string
			Variables map[string]any
		}
		if json.NewDecoder(req.Body).Decode(&q) != nil {
			t.Fatal("invalid request")
		}
		if !strings.Contains(q.Query, "locatorVersion:2") {
			t.Error("missing explicit v2")
		}
		if check != nil {
			check(q.Query, q.Variables)
		}
		raw, _ := json.Marshal(map[string]any{"data": map[string]any{field: payload}})
		return response(200, string(raw)), nil
	})}
}

func locatorV2Ref(componentType string, team any) map[string]any {
	return map[string]any{"schemaVersion": "riido-work-reference.v2", "workspaceId": "workspace-one", "teamId": team, "id": "work-one", "key": "SC-1", "keyNumber": "1", "componentType": componentType}
}

func TestWorkspaceLocatorV2Reference(t *testing.T) {
	for _, target := range []struct {
		name, componentType string
		team                any
	}{{"project", "project", nil}, {"milestone", "milestone", nil}, {"task", "task", "team-one"}, {"subtask", "subtask", "team-one"}} {
		for _, ref := range []string{"SC-1", "work-one"} {
			t.Run(target.name+"/"+ref, func(t *testing.T) {
				r := locatorV2Reader(t, "resolveCLIWorkReference", locatorV2Ref(target.componentType, target.team), func(_ string, v map[string]any) {
					if v["workspaceId"] != "workspace-one" || v["reference"] != ref {
						t.Error("lost scope/reference", v)
					}
				})
				got, err := r.ResolveWorkReference(t.Context(), WorkReferenceRequest{Reference: ref, WorkspaceID: "workspace-one", LocatorVersion: 2})
				if err != nil || !got.Locator().valid() {
					t.Fatal(got, err)
				}
			})
		}
	}
	for _, change := range []struct {
		name  string
		apply func(map[string]any)
	}{
		{"missing-team", func(m map[string]any) { delete(m, "teamId") }},
		{"empty-team", func(m map[string]any) { m["teamId"] = "" }},
		{"fake-team", func(m map[string]any) { m["teamId"] = "team-one" }},
		{"wrong-workspace", func(m map[string]any) { m["workspaceId"] = "another-workspace" }},
		{"wrong-version", func(m map[string]any) { m["schemaVersion"] = "riido-work-reference.v1" }},
		{"task-null-team", func(m map[string]any) { m["componentType"] = "task" }},
		{"wrong-id", func(m map[string]any) { m["id"] = "other-work" }},
	} {
		t.Run(change.name, func(t *testing.T) {
			p := locatorV2Ref("project", nil)
			change.apply(p)
			r := locatorV2Reader(t, "resolveCLIWorkReference", p, nil)
			if _, err := r.ResolveWorkReference(t.Context(), WorkReferenceRequest{Reference: "work-one", WorkspaceID: "workspace-one", LocatorVersion: 2}); !errors.Is(err, ErrInvalidResponse) {
				t.Fatal("accepted invalid locator", err)
			}
		})
	}
}

func TestWorkspaceLocatorV2Search(t *testing.T) {
	for _, tc := range []struct {
		name, kind, componentType string
		team                      any
		valid                     bool
	}{{"project-null", "component", "project", nil, true}, {"milestone-null", "component", "milestone", nil, true}, {"task-current", "component", "task", "team-one", true}, {"task-null", "component", "task", nil, false}, {"project-team", "component", "project", "team-one", false}, {"backlog-team", "backlog", "", "team-one", true}, {"intake-null", "intake", "", nil, false}} {
		t.Run(tc.name, func(t *testing.T) {
			item := map[string]any{"kind": tc.kind, "id": "work-one", "teamId": tc.team, "title": "hello", "componentType": nil, "key": nil, "keyNumber": nil, "updatedAt": "2026-11-13T00:00:00Z", "matchedFields": []string{"title"}}
			if tc.kind == "component" {
				item["componentType"], item["key"], item["keyNumber"] = tc.componentType, "SC-1", "1"
			}
			p := map[string]any{"schemaVersion": "riido-work-search.v2", "workspaceId": "workspace-one", "query": "hello", "limit": 20, "complete": true, "nextCursor": nil, "items": []any{item}}
			r := locatorV2Reader(t, "searchWorkspaceWork", p, nil)
			got, err := r.Search(t.Context(), WorkSearchRequest{WorkspaceID: "workspace-one", Query: "hello", LocatorVersion: 2})
			if (err == nil) != tc.valid {
				t.Fatal(got, err)
			}
			if tc.valid && tc.team == nil {
				raw, _ := json.Marshal(got)
				if !strings.Contains(string(raw), `"teamId":null`) {
					t.Fatal("null team lost")
				}
			}
		})
	}
}

func scopeV2Fixture(t *testing.T, kind string, team *string) *ReadScopeV2 {
	t.Helper()
	base := readScopeFixture(t)
	base.SchemaVersion, base.RootID = "riido-read-scope.v2", "work-one"
	w := base.Works[0]
	w.ID, w.Kind = "work-one", kind
	w.Details = json.RawMessage(`{"id":"work-one","componentType":"project"}`)
	if kind == "backlog" {
		w.BacklogDetails = json.RawMessage(`{"id":"work-one"}`)
	}
	if kind == "intake" {
		w.IntakeDetails = json.RawMessage(`{"id":"work-one"}`)
	}
	w.Body.ResourceID = "work-one"
	w.Comments.ResourceID, w.Comments.ResourceKind = "work-one", kind
	body := &ReadBodyPageV2{ReadBodyPage: *w.Body, SchemaVersion: "riido-read-body.v2", WorkspaceID: "workspace-one", TeamID: team, ResourceKind: kind}
	comments := &ReadScopeCommentPageV2{ReadScopeCommentPage: *w.Comments, SchemaVersion: "riido-read-comments.v2", WorkspaceID: "workspace-one", TeamID: team}
	return &ReadScopeV2{ReadScope: *base, WorkspaceID: "workspace-one", TeamID: team, Works: []*ReadScopeWorkV2{{ReadScopeWork: *w, WorkspaceID: "workspace-one", TeamID: team, Body: body, Comments: comments}}}
}

func TestWorkspaceLocatorV2ReadScope(t *testing.T) {
	for _, kind := range []string{"component", "backlog", "intake"} {
		t.Run(kind, func(t *testing.T) {
			var team *string
			componentType := "project"
			if kind != "component" {
				value := "team-one"
				team = &value
				componentType = "task"
			}
			p := scopeV2Fixture(t, kind, team)
			r := locatorV2Reader(t, "getComponentReadScope", p, func(_ string, v map[string]any) {
				if v["workspaceId"] != "workspace-one" {
					t.Fatal("lost workspace")
				}
				if team == nil && v["teamId"] != nil {
					t.Fatal("invented team")
				}
			})
			got, err := r.ContextV2(t.Context(), WorkLocator{ID: "work-one", WorkspaceID: "workspace-one", TeamID: team, ComponentType: componentType})
			if err != nil || got.Works[0].Kind != kind {
				t.Fatal(got, err)
			}
		})
	}
	for _, tc := range []struct {
		name   string
		change func(*ReadScopeV2)
	}{
		{"wrong-workspace", func(s *ReadScopeV2) { s.WorkspaceID = "other-workspace" }},
		{"nested-workspace", func(s *ReadScopeV2) { s.Works[0].Body.WorkspaceID = "other-workspace" }},
		{"nested-kind", func(s *ReadScopeV2) { s.Works[0].Comments.ResourceKind = "backlog" }},
		{"nested-team", func(s *ReadScopeV2) { v := "team-one"; s.Works[0].Body.TeamID = &v }},
		{"nested-schema", func(s *ReadScopeV2) { s.Works[0].Body.SchemaVersion = "riido-read-body.v1" }},
		{"duplicate", func(s *ReadScopeV2) { s.Works = append(s.Works, s.Works[0]) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := scopeV2Fixture(t, "component", nil)
			tc.change(p)
			r := locatorV2Reader(t, "getComponentReadScope", p, nil)
			if _, err := r.ContextV2(t.Context(), WorkLocator{ID: "work-one", WorkspaceID: "workspace-one", ComponentType: "project"}); !errors.Is(err, ErrInvalidResponse) {
				t.Fatal(err)
			}
		})
	}
}

func TestWorkspaceLocatorV2Continuation(t *testing.T) {
	locator := WorkLocator{ID: "work-one", WorkspaceID: "workspace-one", ComponentType: "project"}
	scope := scopeV2Fixture(t, "component", nil)
	q := ReadBodyRequest{ComponentID: locator.ID, Kind: "component", Part: "document", Version: scope.Works[0].Body.Version, Limit: 100}
	r := locatorV2Reader(t, "getComponentReadScopeBody", scope.Works[0].Body, func(_ string, v map[string]any) {
		if v["workspaceId"] != locator.WorkspaceID || v["teamId"] != nil || v["version"] != q.Version {
			t.Fatal(v)
		}
	})
	if _, err := r.BodyV2(t.Context(), locator, q); err != nil {
		t.Fatal(err)
	}
	q.Version = "sha256:" + strings.Repeat("b", 64)
	if _, err := r.BodyV2(t.Context(), locator, q); !errors.Is(err, ErrInvalidResponse) {
		t.Fatal("wrong version accepted", err)
	}
	comments := scope.Works[0].Comments
	cq := ReadScopeCommentsRequest{ResourceID: locator.ID, ResourceKind: "component", Page: 1, Limit: 20}
	r = locatorV2Reader(t, "getReadScopeComments", comments, nil)
	if _, err := r.ScopeCommentsV2(t.Context(), locator, cq); err != nil {
		t.Fatal(err)
	}
	comments.WorkspaceID = "other-workspace"
	if _, err := r.ScopeCommentsV2(t.Context(), locator, cq); !errors.Is(err, ErrInvalidResponse) {
		t.Fatal("cross-workspace page accepted", err)
	}
}

func TestWorkspaceLocatorV2WorkspaceActors(t *testing.T) {
	locator := WorkLocator{ID: "work-one", WorkspaceID: "workspace-one", ComponentType: "project"}
	p := map[string]any{"schemaVersion": "riido-work-detail.v2", "workspaceId": "workspace-one", "ownerTeamId": nil, "canEdit": true, "id": "work-one", "componentType": "project", "title": "workspace project", "componentLabels": []any{}, "createdByWorkspaceMember": map[string]any{"id": "workspace-member", "nickname": "creator", "status": "active"}, "lastUpdatedByWorkspaceMember": nil, "componentMembers": []any{map[string]any{"id": "membership-one", "memberOrder": 0, "teamMember": nil, "workspaceMember": map[string]any{"id": "workspace-member", "nickname": "owner", "status": "active"}}}}
	r := locatorV2Reader(t, "getComponent", p, func(query string, _ map[string]any) {
		for _, field := range []string{"workspaceMember{id nickname status}", "createdByWorkspaceMember{id nickname status}", "lastUpdatedByWorkspaceMember{id nickname status}"} {
			if !strings.Contains(query, field) {
				t.Error("missing workspace actor selection", field)
			}
		}
	})
	got, err := r.WorkV2(t.Context(), locator)
	if err != nil || got.CreatedByWorkspaceMember == nil || got.CreatedByWorkspaceMember.ID != "workspace-member" || len(got.ComponentMembers) != 1 || got.ComponentMembers[0].WorkspaceMember == nil || got.ComponentMembers[0].TeamMember != nil {
		t.Fatal("workspace actor identity lost", got, err)
	}
	s := scopeV2Fixture(t, "component", nil)
	s.Works[0].Comments.Items = []json.RawMessage{json.RawMessage(`{"id":"comment-one","author":null,"workspaceAuthor":{"id":"workspace-member","nickname":"comment author","status":"active"}}`)}
	r = locatorV2Reader(t, "getComponentReadScope", s, func(query string, _ map[string]any) {
		if !strings.Contains(query, "workspaceAuthor{id nickname status}") || !strings.Contains(query, "workspaceUploader{id nickname status}") {
			t.Error("context omitted workspace actors")
		}
	})
	scope, err := r.ContextV2(t.Context(), locator)
	if err != nil || !strings.Contains(string(scope.Works[0].Comments.Items[0]), `"workspaceAuthor"`) {
		t.Fatal("workspace comment actor lost", err)
	}
}

func TestWorkspaceLocatorV2NativeKeys(t *testing.T) {
	for _, tc := range []struct{ name, request, key, number string }{
		{"underscore-current", "WS_PM-123", "WS_PM-123", "123"},
		{"short-underscore-current", "WS_P-1", "WS_P-1", "WS_P-1"},
		{"legacy-alias-current-key", "OLD-11", "NEW_PM-42", "NEW_PM-42"},
		{"underscore-alias-current-key", "OLD_PM-11", "NEW_PM-42", "42"},
		{"lowercase-underscore-alias", "old_pm-11", "NEW_PM-42", "42"},
		{"lowercase-key-shaped-id", "opaque_ref-7", "SPACE-7", "SPACE-7"},
		{"uppercase-key-shadow", "OPAQUE_REF-7", "OPAQUE_REF-7", "OPAQUE_REF-7"},
		{"opaque-id-current-key", "work-one", "NEW_PM-42", "42"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := locatorV2Ref("project", nil)
			p["key"], p["keyNumber"] = tc.key, tc.number
			id := "work-one"
			if tc.name == "lowercase-key-shaped-id" {
				id = tc.request
				p["id"] = id
			}
			r := locatorV2Reader(t, "resolveCLIWorkReference", p, func(query string, v map[string]any) {
				if query != workReferenceV2Query || v["reference"] != tc.request || v["workspaceId"] != "workspace-one" {
					t.Error("changed fixed native reference", v)
				}
			})
			got, err := r.ResolveWorkReference(t.Context(), WorkReferenceRequest{Reference: tc.request, WorkspaceID: "workspace-one", LocatorVersion: 2})
			if err != nil || got.ID != id || got.Key != tc.key || got.KeyNumber != tc.number {
				t.Fatal("native key/alias rejected", got, err)
			}
		})
	}
	for _, tc := range []struct{ name, reference, key, number, id, workspace string }{
		{"inconsistent-canonical-pair", "OLD-11", "NEW_PM-42", "43", "work-one", "workspace-one"},
		{"missing-canonical-key", "OLD-11", "", "42", "work-one", "workspace-one"},
		{"opaque-id-mismatch", "work-one", "NEW_PM-42", "42", "other-work", "workspace-one"},
		{"alias-workspace-mismatch", "OLD-11", "NEW_PM-42", "42", "work-one", "other-workspace"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := locatorV2Ref("project", nil)
			p["key"], p["keyNumber"], p["id"], p["workspaceId"] = tc.key, tc.number, tc.id, tc.workspace
			r := locatorV2Reader(t, "resolveCLIWorkReference", p, nil)
			if _, err := r.ResolveWorkReference(t.Context(), WorkReferenceRequest{Reference: tc.reference, WorkspaceID: "workspace-one", LocatorVersion: 2}); !errors.Is(err, ErrInvalidResponse) {
				t.Fatal("invalid canonical locator accepted", err)
			}
		})
	}
}
