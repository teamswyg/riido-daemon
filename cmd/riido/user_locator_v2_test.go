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
					fmt.Fprint(w, `{"data":{"resolveCLIWorkReference":{"schemaVersion":"riido-work-reference.v2","workspaceId":"workspace-one","teamId":null,"id":"work-one","key":"SC-1","keyNumber":"1","componentType":"project"}}}`)
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
			if calls != 2 || len(h.daemon.requests) != 0 || !strings.Contains(string(payload), `"workspaceId":"workspace-one"`) {
				t.Fatal("lost workspace or wrong request path", calls, h.stdout.String())
			}
			if !strings.Contains(string(payload), `"teamId":null`) && !strings.Contains(string(payload), `"ownerTeamId":null`) {
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
		{"tree", "work-one", "--workspace", "workspace-one", "--locator-version", "2"},
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

func TestUserWorkspaceLocatorV2DeferredHint(t *testing.T) {
	page := userapi.ReadScopeCommentPageV2{ReadScopeCommentPage: userapi.ReadScopeCommentPage{ResourceID: "work-one", ResourceKind: "component", Status: "deferred", Version: "sha256:page", Page: 1}, WorkspaceID: "workspace-one", ContentPage: &userapi.ReadBodyPageV2{ReadBodyPage: userapi.ReadBodyPage{Part: "comments_page", Version: "sha256:body"}}}
	var out bytes.Buffer
	renderScopeCommentsView(&out, cliui.Terminal{Interactive: true, Width: 160}, page.ContentView(), workBodyV2Command(page.ResourceID, page.WorkspaceID, page.ResourceKind, page.TeamID))
	if !strings.Contains(out.String(), `--workspace "workspace-one" --locator-version 2`) || strings.Contains(out.String(), "<TEAM-ID>") || strings.Contains(out.String(), "--team-id") {
		t.Fatal("invalid workspace continuation", out.String())
	}
}
