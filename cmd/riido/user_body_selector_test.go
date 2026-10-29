package main

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func TestUserTypedBodySelectorUsesProtectedSessionWithoutDaemon(t *testing.T) {
	for _, tc := range []struct {
		name, part, field string
		meeting           bool
		flags             []string
		selector          map[string]any
	}{
		{"work_metadata", "metadata", "getComponentReadScopeBody", false, nil, map[string]any{}},
		{"backlog_replies", "comments_page", "getBacklogReadScopeBody", false, []string{"--kind", "backlog", "--comment-id", "parent-one", "--page", "2", "--comments-limit", "3"}, map[string]any{"parentCommentId": "parent-one", "page": float64(2), "limit": float64(3)}},
		{"meeting_board", "board_page", "getMeetingReadScopeBody", true, []string{"--board-cursor", "board-page-two", "--board-limit", "17"}, map[string]any{"boardCursor": "board-page-two", "limit": float64(17)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			version := "sha256:" + strings.Repeat("a", 64)
			calls := 0
			service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				var req struct {
					Query     string
					Variables map[string]any
				}
				if json.NewDecoder(r.Body).Decode(&req) != nil || r.Header.Get("Authorization") != "Bearer fixture-access" || !strings.Contains(req.Query, tc.field) || !reflect.DeepEqual(req.Variables["selector"], tc.selector) || req.Variables["limit"] != float64(64) {
					t.Fatalf("fixed typed selector request was lost: %+v", req)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{tc.field: map[string]any{"resourceId": "source-one", "part": tc.part, "status": "available", "source": "riido_" + tc.part, "format": "json", "version": version, "offset": 1, "content": "}", "complete": true, "nextOffset": nil, "selector": tc.selector}}})
			})
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			args := append([]string{"body", "source-one", "--team-id", "team-one", "--part", tc.part, "--version", version, "--offset", "1", "--limit", "64", "--json"}, tc.flags...)
			var err error
			if tc.meeting {
				err = runMeeting(h.io, args)
			} else {
				err = runWork(h.io, args)
			}
			if err != nil {
				t.Fatal(err)
			}
			if calls != 1 || len(h.daemon.requests) != 0 {
				t.Fatal("typed live read crossed the daemon boundary")
			}
		})
	}
}

func TestUserTypedBodyRejectsMixedSelectorsBeforeCredentials(t *testing.T) {
	for _, flags := range [][]string{
		{"--part", "metadata", "--page", "1"},
		{"--part", "document", "--comments-limit", "20"},
		{"--part", "comments_page", "--board-limit", "1"},
		{"--part", "comments_page", "--comments-limit", "101"},
		{"--part", "comments_page", "--page", "2147483648"},
		{"--part", "comments_page", "--comment-id", ""},
	} {
		h := newContextHarness(t, nil)
		h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
			t.Fatal("invalid selector accessed credentials")
			return nil, nil
		}
		args := append([]string{"body", "source-one", "--team-id", "team-one", "--version", "sha256:" + strings.Repeat("a", 64), "--json"}, flags...)
		if code := exitCodeOf(t, runWork(h.io, args)); code != 2 {
			t.Fatalf("%v exit=%d", flags, code)
		}
	}
}
