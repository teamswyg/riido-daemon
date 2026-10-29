package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

const userCommentFixture = `{"id":"comment-one","status":"active","createdAt":"2026-10-28T00:00:00.000Z","updatedAt":"2026-10-28T00:01:00.000Z","body":"<p>댓글</p>","author":{"id":"member-one","status":"active","workspaceProfile":{"nickname":"작성자","status":"active"}},"replyCount":1,"isFromSlack":true,"slackMessageUrl":"https://slack.example/message","isFromDiscord":false,"discordMessageUrl":null}`

func commentService(t *testing.T, handler http.HandlerFunc) *standalonecli.Service {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	store := &cliMemoryStore{present: true, credential: userauth.Credential{SessionID: "fixture-session", AccessToken: "fixture-access", RefreshToken: "fixture-refresh", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour)}}
	service, err := standalonecli.New(standalonecli.Config{APIOrigin: server.URL, Store: store})
	if err != nil {
		t.Fatal(err)
	}
	return service
}

func TestUserCommentsReadsOffsetPageAfterMinimalTeamCheck(t *testing.T) {
	for _, tc := range []struct {
		name, page, limit string
		flags             []string
	}{
		{"defaults", "1", "20", nil},
		{"explicit", "2", "100", []string{"--page", "2", "--limit", "100"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls []string
			service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					Query     string                     `json:"query"`
					Variables map[string]json.RawMessage `json:"variables"`
				}
				if r.URL.Path != "/graphql" || r.Header.Get("Authorization") != "Bearer fixture-access" || json.NewDecoder(r.Body).Decode(&request) != nil {
					t.Error("expected authenticated GraphQL read")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				if string(request.Variables["componentId"]) != `"work-one"` {
					t.Error("wrong component")
				}
				if strings.Contains(request.Query, "getComponent(") {
					calls = append(calls, "scope")
					if string(request.Variables["teamId"]) != `"team-one"` || strings.Contains(request.Query, "document") || strings.Contains(request.Query, "getDocument") {
						t.Error("team check must not fetch document content")
					}
					fmt.Fprint(w, `{"data":{"getComponent":{"id":"work-one"}}}`)
					return
				}
				calls = append(calls, "comments")
				var page map[string]string
				if json.Unmarshal(request.Variables["pagination"], &page) != nil || page["page"] != tc.page || page["limit"] != tc.limit || !strings.Contains(request.Query, "getComponentComments") {
					t.Error("wrong offset query or variables")
				}
				prev, next, total := "null", "2", "21"
				if tc.page == "2" {
					prev, next, total = "1", "null", "101"
				}
				fmt.Fprintf(w, `{"data":{"getComponentComments":{"comments":[%s],"meta":{"total":%s,"currentPage":%s,"nextPage":%s,"prevPage":%s,"lastPage":2}}}}`, userCommentFixture, total, tc.page, next, prev)
			})
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			args := append([]string{"comments", "work-one", "--team-id", "team-one", "--json"}, tc.flags...)
			if err := runWork(h.io, args); err != nil {
				t.Fatal(err)
			}
			envelope := decodeEnvelope(t, h.stdout.Bytes())
			data := envelope.Data.(map[string]any)
			comments := data["comments"].([]any)
			comment := comments[0].(map[string]any)
			if data["mode"] != "user" || comment["body"] != "<p>댓글</p>" || comment["replyCount"] != float64(1) || comment["author"] == nil || comment["slackMessageUrl"] != "https://slack.example/message" || data["meta"] == nil || envelope.Page != nil {
				t.Fatalf("lost comment fields or invented cursor: %s", h.stdout.String())
			}
			if !reflect.DeepEqual(calls, []string{"scope", "comments"}) || len(h.daemon.requests) != 0 {
				t.Fatalf("wrong read path: %v", calls)
			}
			if strings.Contains(h.stdout.String()+h.stderr.String(), "fixture-access") {
				t.Fatal("credential leaked")
			}
		})
	}
}

func TestUserRepliesPassParentWorkAndTeamWithoutPagination(t *testing.T) {
	for _, body := range []string{`[]`, `[{"id":"reply-one","status":"active","body":"답글","createdAt":"2026-10-28T00:00:00.000Z","updatedAt":"2026-10-28T00:01:00.000Z","author":null,"isFromSlack":false,"slackMessageUrl":null,"isFromDiscord":false,"discordMessageUrl":null}]`} {
		calls := 0
		service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
			calls++
			var request struct {
				Query     string
				Variables map[string]string
			}
			if r.URL.Path != "/graphql" || json.NewDecoder(r.Body).Decode(&request) != nil || !strings.Contains(request.Query, "getCommentReplies") || !reflect.DeepEqual(request.Variables, map[string]string{"componentId": "work-one", "teamId": "team-one", "commentId": "comment-one"}) {
				t.Error("reply query omitted parent, work or team")
			}
			fmt.Fprintf(w, `{"data":{"getCommentReplies":%s}}`, body)
		})
		h := newContextHarness(t, nil)
		h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
		if err := runWork(h.io, []string{"replies", "work-one", "--team-id", "team-one", "--comment-id", "comment-one", "--json"}); err != nil {
			t.Fatal(err)
		}
		envelope := decodeEnvelope(t, h.stdout.Bytes())
		data := envelope.Data.(map[string]any)
		if data["replies"] == nil || data["meta"] != nil || envelope.Page != nil || calls != 1 || len(h.daemon.requests) != 0 {
			t.Fatal("reply response invented a page or lost its array")
		}
	}
}
