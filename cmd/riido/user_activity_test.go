package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

const activityFixture = `{"id":"activity-one","createdAt":"2026-10-29T00:00:00Z","category":"","action":"update","property":null,"oldValue":"","newValue":"{\"제목\":\"한글🙂\"}","author":{"id":"member-one","status":"active","workspaceProfile":{"nickname":"","status":"active","avatar":{"fileLocation":null}}}}`

func activityArgs() []string {
	return []string{"activity", "work-one", "--team-id", "team-one", "--json"}
}

func TestUserActivityReadsOneTypedAuthenticatedPage(t *testing.T) {
	for _, tc := range []struct {
		page, limit string
		flags       []string
	}{
		{"1", "20", nil}, {"2", "100", []string{"--page", "2", "--limit", "100"}},
	} {
		t.Run(tc.page, func(t *testing.T) {
			calls := 0
			meta := `{"total":1,"currentPage":1,"nextPage":null,"prevPage":null,"lastPage":1}`
			if tc.page == "2" {
				meta = `{"total":101,"currentPage":2,"nextPage":null,"prevPage":1,"lastPage":2}`
			}
			service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				var request struct {
					Query     string         `json:"query"`
					Variables map[string]any `json:"variables"`
				}
				want := map[string]any{"componentId": "work-one", "teamId": "team-one", "pagination": map[string]any{"page": tc.page, "limit": tc.limit}}
				if r.URL.Path != "/graphql" || r.Header.Get("Authorization") != "Bearer fixture-access" || json.NewDecoder(r.Body).Decode(&request) != nil || !strings.Contains(request.Query, "getComponentActivityLogs(") || !strings.Contains(request.Query, "ComponentP06ActivityPaginationInput!") || !reflect.DeepEqual(request.Variables, want) {
					t.Error("expected one scoped activity query")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				fmt.Fprint(w, `{"data":{"getComponentActivityLogs":{"activityLogs":[`+activityFixture+`],"meta":`+meta+`}}}`)
			})
			h := newContextHarness(t, nil)
			h.io.userSession = func(args contextcli.Args) (*standalonecli.Service, error) {
				if args.NoInput {
					t.Error("default user read disabled presence")
				}
				return service, nil
			}
			if err := runWork(h.io, append(activityArgs(), tc.flags...)); err != nil {
				t.Fatal(err)
			}
			env := decodeEnvelope(t, h.stdout.Bytes())
			data := env.Data.(map[string]any)
			var wantRows, wantMeta any
			_ = json.Unmarshal([]byte("["+activityFixture+"]"), &wantRows)
			_ = json.Unmarshal([]byte(meta), &wantMeta)
			if calls != 1 || len(h.daemon.requests) != 0 || data["mode"] != "user" || !reflect.DeepEqual(data["activityLogs"], wantRows) || !reflect.DeepEqual(data["meta"], wantMeta) || env.Page != nil || len(env.Sources) != 0 || len(env.Warnings) != 0 {
				t.Fatalf("activity data changed: %s", h.stdout.String())
			}
			if strings.Contains(h.stdout.String()+h.stderr.String(), "fixture-access") {
				t.Fatal("credential leaked")
			}
		})
	}
}

func TestUserActivityHelpAndModeBoundary(t *testing.T) {
	h := newContextHarness(t, nil)
	if err := runWork(h.io, []string{"activity", "--help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(h.stderr.String(), "work activity <ID> --team-id <TEAM-ID>") {
		t.Fatal("user activity help missing")
	}
	commands := contextcli.Capabilities("test").Commands
	i := slices.IndexFunc(commands, func(c contextcli.CommandCapability) bool { return c.Name == "work activity" })
	if i < 0 || commands[i].Status != contextcli.CommandAvailable || !slices.Equal(commands[i].Modes, []contextcli.Mode{contextcli.ModeUser}) || !contextcli.ReadCommandUnavailable("work activity") {
		t.Fatal("activity modes are inaccurate")
	}
	for _, args := range [][]string{{"activity", "RIID-5900", "--json"}, activityArgs()} {
		h = newContextHarness(t, agentVars())
		h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
			t.Fatal("agent opened user credential")
			return nil, nil
		}
		if code := exitCodeOf(t, runWork(h.io, args)); code != 8 || len(h.daemon.requests) != 0 {
			t.Fatalf("agent activity code=%d calls=%d", code, len(h.daemon.requests))
		}
	}
}
