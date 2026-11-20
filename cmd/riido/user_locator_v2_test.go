package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/cliui"
	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func TestUserWorkspaceLocatorV2Commands(t *testing.T) {
	version := "sha256:" + strings.Repeat("a", 64)
	for _, tc := range []struct {
		name        string
		args        []string
		field, body string
	}{
		{"tree", []string{"tree", "work-one"}, "getComponentReadScope", `{"schemaVersion":"riido-read-scope.v2","workspaceId":"workspace-one","rootId":"work-one","teamId":null,"status":"available","selection":{"childrenLimit":10},"works":[{"id":"work-one","workspaceId":"workspace-one","kind":"component","teamId":null,"reasons":[{"kind":"root"}],"summary":{"id":"work-one","kind":"component","title":"Project","keyNumber":null,"componentType":"project"}}]}`},
		{"relations", []string{"relations", "work-one"}, "getComponentRelatedIssues", `[]`},
		{"activity", []string{"activity", "work-one"}, "getComponentActivityLogs", `{"activityLogs":[],"meta":{"total":0,"currentPage":1,"nextPage":null,"prevPage":null,"lastPage":0}}`},
		{"work", []string{"get", "work-one"}, "getComponent", `{"schemaVersion":"riido-work-detail.v2","workspaceId":"workspace-one","ownerTeamId":null,"canEdit":true,"id":"work-one","componentType":"project","title":"workspace project","componentMembers":[],"componentLabels":[]}`},
		{"body", []string{"body", "work-one", "--version", version}, "getComponentReadScopeBody", `{"schemaVersion":"riido-read-body.v2","workspaceId":"workspace-one","teamId":null,"resourceKind":"component","resourceId":"work-one","part":"document","status":"available","source":"provider_json","format":"json","version":"` + version + `","offset":0,"content":"{}","nextOffset":null,"complete":true,"selector":null}`},
		{"comments", []string{"comments", "work-one"}, "getReadScopeComments", `{"schemaVersion":"riido-read-comments.v2","workspaceId":"workspace-one","teamId":null,"resourceKind":"component","resourceId":"work-one","parentCommentId":null,"limit":20,"version":"` + version + `","status":"empty","page":1,"nextPage":null,"complete":true,"items":[]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				var q struct {
					Query     string
					Variables map[string]any
				}
				if json.NewDecoder(r.Body).Decode(&q) != nil || !strings.Contains(q.Query, "locatorVersion:2") || q.Variables["workspaceId"] != "workspace-one" {
					t.Error("version/workspace lost")
				}
				if strings.Contains(q.Query, "resolveCLIWorkReference") {
					fmt.Fprint(w, `{"data":{"resolveCLIWorkReference":{"schemaVersion":"riido-work-reference.v2","workspaceId":"workspace-one","teamId":null,"id":"work-one","key":"","keyNumber":"","componentType":"project"}}}`)
					return
				}
				if q.Variables["teamId"] != nil || !strings.Contains(q.Query, tc.field+"(") {
					t.Error("wrong continuation target")
				}
				fmt.Fprintf(w, `{"data":{%q:%s}}`, tc.field, tc.body)
			})
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			args := append(append([]string{}, tc.args...), "--workspace", "workspace-one", "--locator-version", "2", "--json")
			if err := runWork(h.io, args); err != nil {
				t.Fatal(err, h.stdout.String())
			}
			env := decodeEnvelope(t, h.stdout.Bytes())
			payload, _ := json.Marshal(env.Data)
			if calls != 2 || len(h.daemon.requests) != 0 || (tc.name != "relations" && tc.name != "activity" && !strings.Contains(string(payload), `"workspaceId":"workspace-one"`)) {
				t.Fatal("lost workspace or wrong request path", calls, h.stdout.String())
			}
			if tc.name != "relations" && tc.name != "activity" && !strings.Contains(string(payload), `"teamId":null`) && !strings.Contains(string(payload), `"ownerTeamId":null`) {
				t.Fatal("null team lost", h.stdout.String())
			}
		})
	}
}

func TestUserWorkspaceLocatorV2RejectsBeforeCredentials(t *testing.T) {
	for _, args := range [][]string{
		{"get", "work-one", "--locator-version", "2"},
		{"get", "work-one", "--workspace", "workspace-one", "--locator-version", "0"},
		{"get", "work-one", "--workspace", "workspace-one", "--locator-version", "3"},
		{"search", "--workspace", "workspace-one", "--query", "hello", "--locator-version", "2.0"},
		{"tree", "work-one", "--locator-version", "2"},
	} {
		h := newContextHarness(t, nil)
		opened := 0
		h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
			opened++
			return nil, fmt.Errorf("must not open")
		}
		if err := runWork(h.io, args); err == nil || opened != 0 || len(h.daemon.requests) != 0 {
			t.Fatal("invalid v2 reached credentials", args, err, opened)
		}
	}
}

func TestUserWorkspaceLocatorV2Search(t *testing.T) {
	calls := 0
	service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		var q struct {
			Query     string
			Variables map[string]any
		}
		if json.NewDecoder(r.Body).Decode(&q) != nil || !strings.Contains(q.Query, "locatorVersion:2") || q.Variables["workspaceId"] != "workspace-one" {
			t.Error("missing v2 request")
		}
		fmt.Fprint(w, `{"data":{"searchWorkspaceWork":{"schemaVersion":"riido-work-search.v2","workspaceId":"workspace-one","query":"hello","limit":20,"complete":true,"nextCursor":null,"items":[{"kind":"component","id":"work-one","teamId":null,"title":"hello","componentType":"project","key":"SC-1","keyNumber":"1","updatedAt":"2026-11-13T00:00:00Z","matchedFields":["title"]}]}}}`)
	})
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	if err := runWork(h.io, []string{"search", "--query", "hello", "--workspace", "workspace-one", "--locator-version", "2", "--json"}); err != nil {
		t.Fatal(err, h.stdout.String())
	}
	env := decodeEnvelope(t, h.stdout.Bytes())
	raw, _ := json.Marshal(env.Data)
	if calls != 1 || !strings.Contains(string(raw), `"teamId":null`) {
		t.Fatal("workspace search locator lost", calls, string(raw))
	}
}

func TestUserWorkspaceLocatorV2ActorRendering(t *testing.T) {
	member := userapi.NativeWorkspaceMember{ID: "workspace-member", Nickname: "Workspace author", Status: "active"}
	if workMembers([]userapi.WorkMember{{WorkspaceMember: &member}}) != "Workspace author" {
		t.Fatal("workspace member displayed as unassigned")
	}
	var comment userapi.Comment
	if json.Unmarshal([]byte(`{"id":"comment-one","author":null,"workspaceAuthor":{"id":"workspace-member","nickname":"Workspace author","status":"active"}}`), &comment) != nil || comment.WorkspaceAuthor == nil || comment.Author != nil {
		t.Fatal("workspace actor converted to team actor")
	}
}

func TestUserWorkspaceLocatorV2ProjectPresentation(t *testing.T) {
	lead := "workspace-member"
	work := userapi.WorkV2{Work: userapi.Work{ID: "project-one", Title: "Project", ComponentType: "project", KeyNumber: "OLD-1"}, LeadMemberID: &lead, TeamIDs: []string{"participant-team"}, TaskTeamIDs: []string{"owner-team"}}
	var out bytes.Buffer
	renderWorkDetailView(&out, cliui.Terminal{Width: 100}, work.Work, &work)
	if strings.Contains(out.String(), "OLD-1") || !strings.Contains(out.String(), "participant-team") || !strings.Contains(out.String(), "owner-team") || !strings.Contains(out.String(), lead) {
		t.Fatal(out.String())
	}
	work.ComponentType = "milestone"
	out.Reset()
	renderWorkDetailView(&out, cliui.Terminal{Width: 100}, work.Work, &work)
	if strings.Contains(out.String(), "담당자") || strings.Contains(out.String(), "OLD-1") || strings.Contains(out.String(), "participant-team") {
		t.Fatal("goal invented a member/team", out.String())
	}
}

func TestUserWorkspaceLocatorV2DeferredHint(t *testing.T) {
	page := userapi.ReadScopeCommentPageV2{ReadScopeCommentPage: userapi.ReadScopeCommentPage{ResourceID: "work-one", ResourceKind: "component", Status: "deferred", Version: "sha256:page", Page: 1}, WorkspaceID: "workspace-one", ContentPage: &userapi.ReadBodyPageV2{ReadBodyPage: userapi.ReadBodyPage{Part: "comments_page", Version: "sha256:body"}}}
	var out bytes.Buffer
	renderScopeCommentsView(&out, cliui.Terminal{Interactive: true, Width: 160}, page.ContentView(), workBodyV2Command(page.ResourceID, page.WorkspaceID, page.ResourceKind, page.TeamID))
	if !strings.Contains(out.String(), `--workspace "workspace-one" --locator-version 2`) || strings.Contains(out.String(), "<TEAM-ID>") || strings.Contains(out.String(), "--team-id") {
		t.Fatal("invalid workspace continuation", out.String())
	}
}

func TestUserWorkspaceLocatorV2NativeKeys(t *testing.T) {
	for _, ref := range []string{"ws_pm-123", "WS_P-1", "OLD-11"} {
		t.Run(ref, func(t *testing.T) {
			calls := 0
			service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				var q struct {
					Query     string
					Variables map[string]any
				}
				if json.NewDecoder(r.Body).Decode(&q) != nil {
					t.Fatal("invalid request")
				}
				if strings.Contains(q.Query, "resolveCLIWorkReference") {
					if q.Variables["reference"] != ref || q.Variables["workspaceId"] != "workspace-one" || !strings.Contains(q.Query, "locatorVersion:2") {
						t.Error("native key identity/case lost", q.Variables)
					}
					fmt.Fprint(w, `{"data":{"resolveCLIWorkReference":{"schemaVersion":"riido-work-reference.v2","workspaceId":"workspace-one","teamId":null,"id":"work-one","key":"NEW_PM-42","keyNumber":"42","componentType":"project"}}}`)
					return
				}
				if q.Variables["componentId"] != "work-one" || q.Variables["teamId"] != nil {
					t.Error("canonical locator not used")
				}
				fmt.Fprint(w, `{"data":{"getComponent":{"schemaVersion":"riido-work-detail.v2","workspaceId":"workspace-one","ownerTeamId":null,"canEdit":true,"id":"work-one","componentType":"project","title":"canonical alias target","componentMembers":[],"componentLabels":[]}}}`)
			})
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			if err := runWork(h.io, []string{"get", ref, "--workspace", "workspace-one", "--locator-version", "2", "--json"}); err != nil || calls != 2 {
				t.Fatal("native key did not reach canonical target", err, calls)
			}
		})
	}
}

func TestUserContextProjectKeysStayVisibleInV1(t *testing.T) {
	for _, kind := range []string{"project", "milestone"} {
		for _, summaryOnly := range []bool{false, true} {
			for _, version := range []string{"riido-read-scope.v1", "riido-read-scope.v2"} {
				t.Run(fmt.Sprintf("%s/%s/summary=%t", version, kind, summaryOnly), func(t *testing.T) {
					key := "OLD-42"
					work := &userapi.ReadScopeWork{ID: "work-one", Kind: "component"}
					if summaryOnly {
						work.Summary = &userapi.ReadResourceSummary{ID: "work-one", Kind: "component", Title: "Workspace work", KeyNumber: &key, ComponentType: &kind}
					} else {
						work.Details = json.RawMessage(fmt.Sprintf(`{"id":"work-one","title":"Workspace work","componentType":%q,"keyNumber":"OLD-42"}`, kind))
					}
					scope := userapi.ReadScope{SchemaVersion: version, Works: []*userapi.ReadScopeWork{work}}
					var out bytes.Buffer
					renderContextView(&out, cliui.Terminal{Width: 100}, scope)
					if strings.Contains(out.String(), key) != (version == "riido-read-scope.v1") {
						t.Fatal(out.String())
					}
				})
			}
		}
	}
}
