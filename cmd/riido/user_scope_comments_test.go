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

func TestUserScopeCommentsUsesVersionBoundTypedSourceAndReplyParent(t *testing.T) {
	for _, tc := range []struct{ sub, kind, parent string }{{"comments", "backlog", ""}, {"comments", "intake", ""}, {"replies", "component", "parent-one"}} {
		t.Run(tc.sub+tc.kind, func(t *testing.T) {
			version := "sha256:" + strings.Repeat("a", 64)
			calls := 0
			service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				var req struct {
					Query     string
					Variables map[string]any
				}
				if json.NewDecoder(r.Body).Decode(&req) != nil || !strings.Contains(req.Query, "getReadScopeComments(") || req.Variables["resourceId"] != "work-one" || req.Variables["resourceKind"] != tc.kind || req.Variables["teamId"] != "team-one" || req.Variables["page"] != float64(2) || req.Variables["limit"] != float64(20) || req.Variables["version"] != version {
					t.Fatal("wrong fixed source or continuation request")
				}
				if tc.parent != "" && req.Variables["parentCommentId"] != tc.parent {
					t.Fatal("reply parent was lost")
				}
				var parent any
				if tc.parent != "" {
					parent = tc.parent
				}
				out := map[string]any{"resourceId": "work-one", "resourceKind": tc.kind, "parentCommentId": parent, "version": version, "limit": 20, "status": "available", "page": 2, "nextPage": nil, "complete": true, "items": []any{map[string]any{"id": "comment-one", "body": "댓글 원문", "reactions": []any{}}}}
				if err := json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"getReadScopeComments": out}}); err != nil {
					t.Fatal(err)
				}
			})
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			args := []string{tc.sub, "work-one", "--kind", tc.kind, "--team-id", "team-one", "--page", "2", "--version", version, "--json"}
			if tc.parent != "" {
				args = append(args, "--comment-id", tc.parent)
			}
			if err := runWork(h.io, args); err != nil {
				t.Fatal(err)
			}
			data := decodeEnvelope(t, h.stdout.Bytes()).Data.(map[string]any)
			if data["comment_page"].(map[string]any)["complete"] != true || calls != 1 || len(h.daemon.requests) != 0 {
				t.Fatal("lost typed page or user boundary")
			}
		})
	}
}

func TestUserScopeCommentsRejectsUnversionedContinuationAndSanitizesConflict(t *testing.T) {
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
		t.Fatal("unversioned continuation opened credentials")
		return nil, nil
	}
	args := []string{"comments", "work-one", "--kind", "backlog", "--team-id", "team-one", "--page", "2", "--json"}
	if code := exitCodeOf(t, runWork(h.io, args)); code != 2 {
		t.Fatalf("missing version exit=%d", code)
	}
	service := commentService(t, func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"errors":[{"message":"private-comment-body","extensions":{"code":"conflict"}}]}`)
	})
	h = newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	args = append(args, "--version", "sha256:"+strings.Repeat("a", 64))
	if code := exitCodeOf(t, runWork(h.io, args)); code != 10 || strings.Contains(h.stdout.String(), "private-comment-body") {
		t.Fatalf("conflict was not sanitized: exit=%d", code)
	}
}

func TestUserWorkCommentsRejectMeetingKind(t *testing.T) {
	for _, sub := range []string{"comments", "replies"} {
		h := newContextHarness(t, nil)
		h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
			t.Fatal("work command accepted a meeting kind")
			return nil, nil
		}
		args := []string{sub, "meeting-one", "--kind", "meeting", "--team-id", "team-one", "--json"}
		if sub == "replies" {
			args = append(args, "--comment-id", "parent-one")
		}
		if code := exitCodeOf(t, runWork(h.io, args)); code != 2 {
			t.Fatalf("meeting kind accepted by work: exit=%d", code)
		}
	}
}
