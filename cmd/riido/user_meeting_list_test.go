package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

const (
	userMeetingListFixture    = `{"data":{"getComponentReadScope":{"schemaVersion":"riido-read-scope.v1","rootId":"work-one","teamId":"team-one","status":"available","selection":{"relatedMeetingLimit":5},"meetings":[{"id":"meeting-one","teamId":"other-team","reasons":[{"kind":"meeting","relation":null,"source":null}]}]}}}`
	userMeetingSummaryFixture = `{"data":{"getMeetingReadScope":{"id":"meeting-one","teamId":"other-team","status":"available","summary":{"id":"meeting-one","kind":"meeting","title":"현재 미팅\u001b[31m\n위조 줄"}}}}`
)

func TestUserMeetingListAuthenticatedHTTPAndEscapedText(t *testing.T) {
	for _, asJSON := range []bool{false, true} {
		t.Run(fmt.Sprint(asJSON), func(t *testing.T) {
			var calls atomic.Int32
			service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				var body struct {
					Query     string
					Variables map[string]string
				}
				if r.Header.Get("Authorization") != "Bearer fixture-access" || r.URL.Path != "/graphql" || json.NewDecoder(r.Body).Decode(&body) != nil {
					t.Error("missing authenticated GraphQL request")
				}
				for _, forbidden := range []string{"body{", "board{", "details{", "works{", "contentPage", "mutation"} {
					if strings.Contains(body.Query, forbidden) {
						t.Error("requested body/detail instead of summary")
					}
				}
				if strings.Contains(body.Query, "query RiidoCLIMeetingList(") {
					if body.Variables["componentId"] != "work-one" || body.Variables["teamId"] != "team-one" {
						t.Error("wrong root")
					}
					fmt.Fprint(w, userMeetingListFixture)
				} else {
					if !strings.Contains(body.Query, "query RiidoCLIMeetingSummary(") || body.Variables["meetingId"] != "meeting-one" || body.Variables["teamId"] != "other-team" {
						t.Error("changed meeting identity or cross-team selection")
					}
					fmt.Fprint(w, userMeetingSummaryFixture)
				}
			})
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			args := []string{"list", "--work", "work-one", "--team-id", "team-one"}
			if asJSON {
				args = append(args, "--json")
			}
			if err := runMeeting(h.io, args); err != nil {
				t.Fatal(err)
			}
			if calls.Load() != 2 || len(h.daemon.requests) != 0 || strings.ContainsRune(h.stdout.String(), '\x1b') || strings.Contains(h.stdout.String(), "fixture-access") {
				t.Fatal("wrong requests or unsafe terminal output")
			}
			if asJSON {
				env := decodeEnvelope(t, h.stdout.Bytes())
				raw, _ := json.Marshal(env.Data)
				var data struct {
					Mode        string
					Live        bool
					MeetingList userapi.MeetingList
				}
				if json.Unmarshal(raw, &data) != nil || data.Mode != "user" || !data.Live || data.MeetingList.SelectionLimit != 5 || len(data.MeetingList.Meetings) != 1 || data.MeetingList.Meetings[0].Title != "현재 미팅\x1b[31m\n위조 줄" {
					t.Fatal("wrong user summary envelope")
				}
			} else if !strings.Contains(h.stdout.String(), "선택된 연관 미팅: 1개 (최대 5개, 한 단계)") || strings.Contains(h.stdout.String(), "\n위조 줄") {
				t.Fatal("misleading or unescaped text output")
			}
		})
	}
}

func TestUserMeetingListPartialAndSummaryFailure(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		exit   int
	}{{"partial", http.StatusOK, 0}, {"revoked", http.StatusForbidden, 4}, {"rate", http.StatusTooManyRequests, 7}} {
		t.Run(tc.name, func(t *testing.T) {
			var calls atomic.Int32
			service := commentService(t, func(w http.ResponseWriter, _ *http.Request) {
				if calls.Add(1) == 1 {
					fmt.Fprint(w, userMeetingListFixture)
					return
				}
				w.WriteHeader(tc.status)
				fmt.Fprint(w, strings.Replace(userMeetingSummaryFixture, `"available"`, `"partial"`, 1))
			})
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			err := runMeeting(h.io, []string{"list", "--work", "work-one", "--team-id", "team-one", "--json"})
			if tc.exit == 0 && err != nil || tc.exit != 0 && exitCodeOf(t, err) != tc.exit {
				t.Fatal("wrong command result", err)
			}
			env := decodeEnvelope(t, h.stdout.Bytes())
			if tc.exit == 0 {
				if len(env.Warnings) != 1 || env.Warnings[0].Code != "CONTEXT_INCOMPLETE" {
					t.Fatal("partial context warning absent")
				}
			} else if env.Data != nil || env.Error == nil {
				t.Fatal("failed summary returned a partial list")
			}
			if calls.Load() != 2 || len(h.daemon.requests) != 0 {
				t.Fatal("unexpected retry or daemon fallback")
			}
		})
	}
}

func TestUserMeetingListRejectsScopeOverridesBeforeCredentials(t *testing.T) {
	for _, flags := range [][]string{{"--workspace", "x"}, {"--cursor", "x"}, {"--page", "2"}, {"--budget", "1000"}, {"--limit", "20"}, {"--profile", "x"}, {"--team", "x"}} {
		t.Run(flags[0], func(t *testing.T) {
			h := newContextHarness(t, nil)
			opened := 0
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { opened++; return nil, userapi.ErrAuthentication }
			args := append([]string{"list", "--work", "work-one", "--team-id", "team-one", "--json"}, flags...)
			if exitCodeOf(t, runMeeting(h.io, args)) != 8 || opened != 0 || len(h.daemon.requests) != 0 {
				t.Fatal("unsupported input accessed credentials or daemon")
			}
		})
	}
	for _, tc := range []struct {
		args []string
		exit int
	}{{[]string{"list", "--work", "work-one"}, 2}, {[]string{"list", "--work", "RIID-5972", "--team-id", "team-one"}, 8}} {
		h := newContextHarness(t, nil)
		if exitCodeOf(t, runMeeting(h.io, append(tc.args, "--json"))) != tc.exit {
			t.Fatal("missing IDs or display key accepted")
		}
	}
}
