package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func TestUserWorkReferenceCommandTargetsAndCurrentAccess(t *testing.T) {
	version := "sha256:" + strings.Repeat("a", 64)
	commands := [][]string{{"work", "get", "SC-1"}, {"work", "tree", "SC-1"}, {"work", "relations", "SC-1"}, {"work", "comments", "SC-1"}, {"work", "replies", "SC-1", "--comment-id", "parent-one"}, {"work", "activity", "SC-1"}, {"work", "body", "SC-1", "--version", version}, {"work", "comments", "SC-1", "--kind", "component"}, {"work", "replies", "SC-1", "--kind", "component", "--comment-id", "parent-one"}, {"context", "get", "--work", "SC-1"}, {"resource", "list", "--work", "SC-1"}, {"repo", "list", "--work", "SC-1"}, {"meeting", "list", "--work", "SC-1"}}
	for _, cmd := range commands {
		for _, scope := range []string{"team", "workspace", "both", "default"} {
			t.Run(strings.Join(cmd[:2], "-")+"/"+scope+strings.Join(cmd[3:], ""), func(t *testing.T) {
				calls := 0
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					calls++
					var q struct {
						Query     string
						Variables map[string]json.RawMessage
					}
					if json.NewDecoder(r.Body).Decode(&q) != nil {
						t.Error("invalid request")
						return
					}
					if strings.Contains(q.Query, "resolveCLIWorkReference") {
						if calls != 1 || string(q.Variables["reference"]) != `"SC-1"` {
							t.Error("wrong resolver invocation")
						}
						if scope != "workspace" && scope != "default" && string(q.Variables["teamId"]) != `"team-one"` {
							t.Error("team not intersected")
						}
						if scope != "team" && string(q.Variables["workspaceId"]) != `"workspace-one"` {
							t.Error("workspace not intersected")
						}
						fmt.Fprint(w, `{"data":{"resolveCLIWorkReference":{"schemaVersion":"riido-work-reference.v1","workspaceId":"workspace-one","teamId":"team-one","id":"work-one","key":"SC-1","keyNumber":"1","componentType":"task"}}}`)
						return
					}
					if calls != 2 || string(q.Variables["teamId"]) != `"team-one"` {
						t.Error("resolved team not used", q.Variables)
					}
					id := q.Variables["componentId"]
					if id == nil {
						id = q.Variables["resourceId"]
					}
					if string(id) != `"work-one"` {
						t.Error("resolved ID not used", q.Variables)
					}
					w.WriteHeader(http.StatusForbidden)
				}))
				t.Cleanup(server.Close)
				store := &cliMemoryStore{present: true, credential: userauth.Credential{SessionID: "key-session", AccessToken: "key-access", RefreshToken: "key-refresh", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour)}}
				service, err := standalonecli.New(standalonecli.Config{APIOrigin: server.URL, Store: store, HTTPClient: server.Client()})
				if err != nil {
					t.Fatal(err)
				}
				vars := map[string]string{}
				args := append([]string{}, cmd...)
				if scope == "team" || scope == "both" {
					args = append(args, "--team-id", "team-one")
				}
				if scope == "workspace" || scope == "both" {
					args = append(args, "--workspace", "workspace-one")
				}
				if scope == "default" {
					vars["RIIDO_CLI_WORKSPACE"] = "workspace-one"
				}
				h := newContextHarness(t, vars)
				h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
				err = runKeyCommand(h.io, append(args, "--json"))
				if exitCodeOf(t, err) != 4 || calls != 2 || decodeEnvelope(t, h.stdout.Bytes()).Data != nil || len(h.daemon.requests) != 0 {
					t.Fatalf("current access bypassed or no resolved read: calls=%d output=%s", calls, h.stdout.String())
				}
			})
		}
	}
}

func runKeyCommand(c contextIO, args []string) error {
	switch args[0] {
	case "work":
		return runWork(c, args[1:])
	case "context":
		return runContext(c, args[1:])
	case "resource":
		return runResource(c, args[1:])
	case "repo":
		return runRepo(c, args[1:])
	default:
		return runMeeting(c, args[1:])
	}
}

func TestUserWorkReferenceInvalidPageBeforeStorage(t *testing.T) {
	for _, args := range [][]string{{"work", "body", "SC-1", "--workspace", "workspace-one"}, {"work", "body", "SC-1", "--kind", "backlog", "--workspace", "workspace-one"}, {"work", "comments", "SC-1", "--workspace", "workspace-one", "--page", "0"}, {"work", "comments", "SC-1", "--workspace", "workspace-one", "--kind", "component", "--page", "2"}, {"work", "replies", "SC-1", "--workspace", "workspace-one", "--comment-id", "SC-2"}, {"work", "get", "work-one", "--team-id", "team-one", "--workspace", "workspace-one"}} {
		h := newContextHarness(t, nil)
		h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
			t.Fatal("invalid request opened credentials")
			return nil, nil
		}
		if err := runKeyCommand(h.io, append(args, "--json")); err == nil {
			t.Fatal("invalid request accepted", args)
		}
	}
}
