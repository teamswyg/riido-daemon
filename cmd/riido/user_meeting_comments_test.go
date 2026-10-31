package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func TestUserMeetingCommentsFixedAPIPagesAndReplies(t *testing.T) {
	for _, parent := range []string{"", "parent-one"} {
		t.Run("parent="+parent, func(t *testing.T) {
			version := "sha256:" + strings.Repeat("a", 64)
			calls := 0
			service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				var req struct {
					Query     string
					Variables map[string]any
				}
				if json.NewDecoder(r.Body).Decode(&req) != nil || !strings.Contains(req.Query, "getReadScopeComments(") || req.Variables["resourceKind"] != "meeting" || req.Variables["resourceId"] != "meeting-one" || req.Variables["teamId"] != "team-one" || req.Variables["page"] != float64(2) || req.Variables["limit"] != float64(3) || req.Variables["version"] != version {
					t.Error("fixed meeting query mismatch")
				}
				var p any
				if parent != "" {
					p = parent
					if req.Variables["parentCommentId"] != parent {
						t.Error("parent lost")
					}
				}
				out := map[string]any{"resourceId": "meeting-one", "resourceKind": "meeting", "parentCommentId": p, "version": version, "limit": 3, "status": "available", "page": 2, "nextPage": nil, "complete": true, "items": []any{map[string]any{"id": "comment-one", "body": "미팅 댓글", "reactions": []any{}}}}
				_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"getReadScopeComments": out}})
			})
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			args := []string{"comments", "meeting-one", "--team-id", "team-one", "--page", "2", "--limit", "3", "--version", version, "--json"}
			if parent != "" {
				args = append(args, "--comment-id", parent)
			}
			if err := runMeeting(h.io, args); err != nil {
				t.Fatal(err)
			}
			env := decodeEnvelope(t, h.stdout.Bytes())
			page := env.Data.(map[string]any)["comment_page"].(map[string]any)
			if page["resourceKind"] != "meeting" || page["complete"] != true || calls != 1 || len(h.daemon.requests) != 0 {
				t.Fatal("meeting source or user boundary lost")
			}
		})
	}
}

func TestUserMeetingCommentsRejectBeforeCredentials(t *testing.T) {
	for _, args := range [][]string{
		{"comments"},
		{"comments", "meeting-one"},
		{"comments", "meeting-one", "--team-id", "team-one", "--page", "2"},
		{"comments", "meeting-one", "--team-id", "team-one", "--limit", "101"},
		{"comments", "meeting-one", "--team-id", "team-one", "--page", "0"},
		{"comments", "meeting-one", "--team-id", "team-one", "--kind", "component"},
		{"comments", "meeting-one", "--team-id", "team-one", "--comment-id", ""},
		{"comments", "meeting-one", "--team-id", "team-one", "--version", ""},
		{"comments", "meeting-one", "--team-id", "team-one", "--workspace", "workspace-one"},
		{"comments", "meeting-one", "--team-id", "team-one", "--cursor", "snapshot"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
				t.Error("invalid args opened credentials")
				return nil, nil
			}
			if err := runMeeting(h.io, append(args, "--json")); err == nil {
				t.Fatal("invalid args accepted")
			}
			if len(h.daemon.requests) != 0 {
				t.Fatal("user command used daemon")
			}
		})
	}
	h := newContextHarness(t, map[string]string{"RIIDO_RUN_TOKEN": "fixture"})
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
		t.Error("agent opened user credentials")
		return nil, nil
	}
	if code := exitCodeOf(t, runMeeting(h.io, []string{"comments", "meeting-one", "--team-id", "team-one", "--json"})); code != 8 {
		t.Fatal("agent meeting comments must be unsupported", code)
	}
	if len(h.daemon.requests) != 0 {
		t.Fatal("agent unsupported command used IPC")
	}
}

func TestUserMeetingCommentsSanitizesPermissionAndConflict(t *testing.T) {
	for _, tc := range []struct {
		code string
		exit int
	}{{"permission_denied", 4}, {"conflict", 10}} {
		t.Run(tc.code, func(t *testing.T) {
			service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(map[string]any{"errors": []any{map[string]any{"message": "private-comment-details", "extensions": map[string]string{"code": tc.code}}}})
			})
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			if code := exitCodeOf(t, runMeeting(h.io, []string{"comments", "meeting-one", "--team-id", "team-one", "--json"})); code != tc.exit {
				t.Fatal(code)
			}
			env := decodeEnvelope(t, h.stdout.Bytes())
			if env.Data != nil || strings.Contains(h.stdout.String(), "private-comment-details") {
				t.Fatal("upstream data leaked")
			}
		})
	}
}
