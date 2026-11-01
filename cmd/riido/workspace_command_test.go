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

func TestWorkspaceRootHelp(t *testing.T) {
	if err := run([]string{"workspace", "--help"}); err != nil {
		t.Fatal(err)
	}
}

const userWorkspaceFixture = `{"data":{"getCLIWorkspaces":{"schemaVersion":"riido-cli-workspaces.v1","userId":"user-one","limit":1,"items":[{"id":"ws-one","name":"현재 이름\u001b[31m\n위조 줄","isDemo":false}],"nextCursor":"bmV4dA","complete":false}}}`

func TestWorkspaceCLIAuthenticatedHTTPAndOutput(t *testing.T) {
	for _, asJSON := range []bool{false, true} {
		t.Run(fmt.Sprint(asJSON), func(t *testing.T) {
			var calls atomic.Int32
			service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				var body struct {
					Query     string
					Variables map[string]any
				}
				if r.Header.Get("Authorization") != "Bearer fixture-access" || r.URL.Path != "/graphql" || json.NewDecoder(r.Body).Decode(&body) != nil || !strings.Contains(body.Query, "getCLIWorkspaces(") || strings.Contains(body.Query, "mutation") || body.Variables["limit"] != float64(1) || body.Variables["cursor"] != "cHJldmlvdXM" || len(body.Variables) != 2 {
					t.Error("wrong authenticated read")
				}
				fmt.Fprint(w, userWorkspaceFixture)
			})
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			args := []string{"list", "--limit", "1", "--cursor", "cHJldmlvdXM"}
			if asJSON {
				args = append(args, "--json")
			}
			if err := runWorkspace(h.io, args); err != nil {
				t.Fatal(err)
			}
			if calls.Load() != 1 || len(h.daemon.requests) != 0 || strings.ContainsRune(h.stdout.String(), '\x1b') || strings.Contains(h.stdout.String(), "fixture-access") || strings.Contains(h.stdout.String(), "\n위조 줄") {
				t.Fatal("unsafe output or extra requests")
			}
			if asJSON {
				env := decodeEnvelope(t, h.stdout.Bytes())
				raw, _ := json.Marshal(env.Data)
				var data struct {
					Mode       string
					Live       bool
					Workspaces userapi.WorkspaceList
				}
				if json.Unmarshal(raw, &data) != nil || data.Mode != "user" || !data.Live || data.Workspaces.Complete || len(data.Workspaces.Items) != 1 || data.Workspaces.Items[0].Name != "현재 이름\x1b[31m\n위조 줄" {
					t.Fatal("wrong envelope")
				}
			} else if !strings.Contains(h.stdout.String(), "이번 페이지 1개") || !strings.Contains(h.stdout.String(), `riido workspace list --limit 1 --cursor "bmV4dA"`) {
				t.Fatal("missing pagination guidance")
			}
		})
	}
}

func TestWorkspaceCLIRejectsBeforeCredentials(t *testing.T) {
	for _, tc := range []struct {
		args []string
		exit int
	}{
		{[]string{"list", "--limit", "0"}, 2},
		{[]string{"list", "--limit", "101"}, 2},
		{[]string{"list", "--limit", "abc"}, 2},
		{[]string{"list", "--cursor", ""}, 2},
		{[]string{"list", "--cursor", "bad="}, 2},
		{[]string{"list", "unexpected"}, 2},
		{[]string{"list", "--workspace", "ws"}, 8},
		{[]string{"list", "--profile", "unknown"}, 2},
		{[]string{"list", "--team", "t"}, 8},
		{[]string{"list", "--token", "synthetic-secret"}, 2},
		{[]string{"list", "--api-url", "https://attacker.invalid"}, 2},
	} {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
				t.Error("opened credentials")
				return nil, userapi.ErrAuthentication
			}
			if exitCodeOf(t, runWorkspace(h.io, append(tc.args, "--json"))) != tc.exit || len(h.daemon.requests) != 0 {
				t.Fatal("wrong rejection")
			}
			if strings.Contains(h.stdout.String(), "synthetic-secret") {
				t.Fatal("secret printed")
			}
		})
	}
	h := newContextHarness(t, map[string]string{"RIIDO_RUN_TOKEN": "fixture"})
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
		t.Error("agent opened user credentials")
		return nil, userapi.ErrAuthentication
	}
	if exitCodeOf(t, runWorkspace(h.io, []string{"list", "--json"})) != 8 || len(h.daemon.requests) != 0 {
		t.Fatal("agent mode not blocked")
	}
}

func TestWorkspaceCLIFailureNoPartialResult(t *testing.T) {
	for _, tc := range []struct{ status, exit int }{{403, 4}, {429, 7}, {503, 6}} {
		t.Run(fmt.Sprint(tc.status), func(t *testing.T) {
			var calls atomic.Int32
			service := commentService(t, func(w http.ResponseWriter, _ *http.Request) {
				calls.Add(1)
				w.WriteHeader(tc.status)
				fmt.Fprint(w, "private-server-error")
			})
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			if exitCodeOf(t, runWorkspace(h.io, []string{"list", "--json"})) != tc.exit {
				t.Fatal("wrong exit")
			}
			env := decodeEnvelope(t, h.stdout.Bytes())
			if env.Data != nil || env.Error == nil || calls.Load() != 1 || strings.Contains(h.stdout.String(), "private-server-error") {
				t.Fatal("failed result leaked")
			}
		})
	}
}
