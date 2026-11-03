package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/cliconfig"
	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func cpWorkKeyFixture(t *testing.T, kind string, lookupError ...string) (*standalonecli.Service, *atomic.Int32) {
	t.Helper()
	var calls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		if r.Header.Get("Cookie") != "" {
			t.Error("cookie sent")
		}
		if r.URL.Path == "/graphql" {
			var q struct {
				Query     string
				Variables map[string]string
			}
			if r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer cp-key-access" || json.NewDecoder(r.Body).Decode(&q) != nil || !strings.Contains(q.Query, "resolveCLIWorkReference") || q.Variables["reference"] != "SC-1" || (q.Variables["workspaceId"] != "workspace-one" && q.Variables["teamId"] != "team-one") {
				t.Error("wrong reference request")
			}
			if len(lookupError) == 1 {
				_, _ = fmt.Fprintf(w, `{"data":{"resolveCLIWorkReference":null},"errors":[{"message":"private-reference-canary","extensions":{"code":%q}}]}`, lookupError[0])
				return
			}
			_, _ = io.WriteString(w, `{"data":{"resolveCLIWorkReference":{"schemaVersion":"riido-work-reference.v1","workspaceId":"workspace-one","teamId":"team-one","id":"task-one","key":"SC-1","keyNumber":"1","componentType":"task"}}}`)
			return
		}
		prefix, purpose := "riido_cli_thread_v1_", "task_threads_read"
		if kind == "agent" {
			prefix, purpose = "riido_cli_agent_v1_", "task_agents_read"
		}
		if r.URL.Path == "/oauth/cli/"+kind+"-read-tickets" {
			var q map[string]string
			if r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer cp-key-access" || json.NewDecoder(r.Body).Decode(&q) != nil || len(q) != 2 || q["workspace_id"] != "workspace-one" || q["task_id"] != "task-one" {
				t.Error("wrong resolved ticket scope")
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"schema_version": "riido-cli-" + kind + "-read.v1", "access_token": prefix + strings.Repeat("A", 43), "token_type": "Bearer", "resource": "https://staging.ai-api.riido.io", "purpose": purpose, "expires_at": time.Now().Add(time.Minute).UTC().Format(time.RFC3339Nano)})
			return
		}
		if r.Method != http.MethodGet || r.Header.Get("Authorization") != "Bearer "+prefix+strings.Repeat("A", 43) {
			t.Error("CP original credential or method")
		}
		if kind == "thread" {
			if r.URL.Path != "/v3/client/workspaces/workspace-one/ai-agent/tasks/task-one/threads" {
				t.Error("wrong task")
			}
			_, _ = io.WriteString(w, threadCLIHistory)
		} else {
			if r.URL.Path == "/v1/cli/workspaces/workspace-one/tasks/task-one/agents/agent-one" {
				_, _ = io.WriteString(w, `{"agent":`+userAgentFixture+`,"fetched_at":"2026-11-01T01:00:00Z","policy_checked":false}`)
				return
			}
			if r.URL.Path != "/v1/cli/workspaces/workspace-one/tasks/task-one/agents" {
				t.Error("wrong task")
			}
			cursor := r.URL.Query().Get("cursor")
			if cursor == "" {
				cursor = "0"
			}
			total := 1
			if cursor == "1" {
				total = 51
			}
			_, _ = fmt.Fprintf(w, `{"agents":[%s],"total":%d,"cursor":%q,"fetched_at":"2026-11-01T01:00:00Z","policy_checked":false}`, userAgentFixture, total, cursor)
		}
	}))
	t.Cleanup(server.Close)
	client := &http.Client{Transport: threadCLITransport(func(r *http.Request) (*http.Response, error) {
		want := "monolith.api.riido.io"
		if r.Method == http.MethodGet {
			want = "staging.ai-api.riido.io"
		}
		if r.URL.Host != want || r.URL.Scheme != "https" {
			t.Fatal("wrong canonical target")
		}
		out := r.Clone(r.Context())
		u := *r.URL
		target, _ := url.Parse(server.URL)
		u.Scheme, u.Host = target.Scheme, target.Host
		out.URL, out.Host = &u, target.Host
		return server.Client().Transport.RoundTrip(out)
	})}
	store := &cliMemoryStore{present: true, credential: userauth.Credential{SessionID: "cp-key-session", AccessToken: "cp-key-access", RefreshToken: "cp-key-refresh", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour)}}
	service, err := standalonecli.New(standalonecli.Config{APIOrigin: "https://monolith.api.riido.io", Store: store, HTTPClient: client})
	if err != nil {
		t.Fatal(err)
	}
	return service, &calls
}

func TestUserCPWorkKeys(t *testing.T) {
	for _, kind := range []string{"thread", "agent"} {
		t.Run(kind, func(t *testing.T) {
			service, calls := cpWorkKeyFixture(t, kind)
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			args := []string{"list", "--work", "sc-1", "--workspace", "workspace-one", "--json"}
			var err error
			if kind == "thread" {
				err = runThread(h.io, args)
			} else {
				err = runAgents(h.io, args)
			}
			if exitCodeOf(t, err) != 0 || calls.Load() != 3 {
				t.Fatalf("key rejected or wrong calls=%d: %s", calls.Load(), h.stdout.String())
			}
			if decodeEnvelope(t, h.stdout.Bytes()).Data == nil || len(h.daemon.requests) != 0 {
				t.Fatal("missing result or IPC")
			}
			for _, secret := range []string{"cp-key-access", "cp-key-refresh", "riido_cli_thread_v1_", "riido_cli_agent_v1_"} {
				if strings.Contains(h.stdout.String()+h.stderr.String(), secret) {
					t.Fatal("credential exposed")
				}
			}
		})
	}
}

func TestUserCPWorkKeyScopesAndIDCompatibility(t *testing.T) {
	for _, kind := range []string{"thread", "agent"} {
		for _, mode := range []string{"workspace", "team", "both", "configured", "id-configured", "show", "cursor"} {
			if kind == "thread" && mode == "cursor" {
				continue
			}
			t.Run(kind+"/"+mode, func(t *testing.T) {
				service, calls := cpWorkKeyFixture(t, kind)
				h := newContextHarness(t, nil)
				h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
				args := []string{"list", "--work", "SC-1", "--json"}
				want := int32(3)
				switch mode {
				case "workspace":
					args = append(args, "--workspace", "workspace-one")
				case "team":
					args = append(args, "--team", "team-one")
				case "both":
					args = append(args, "--workspace", "workspace-one", "--team-id", "team-one")
				case "show":
					args = []string{"show", kind + "-one", "--work", "SC-1", "--workspace", "workspace-one", "--json"}
				case "cursor":
					args = append(args, "--workspace", "workspace-one", "--cursor", "1")
				case "configured", "id-configured":
					dir := t.TempDir()
					h.io.configDir = func() (string, error) { return dir, nil }
					if err := (cliconfig.Store{Dir: dir}).Update(context.Background(), func(c *cliconfig.Config) error {
						c.Profiles["production"] = cliconfig.Profile{SelectionOrigin: "https://monolith.riido.io", Workspace: "workspace-one", Team: "team-one"}
						return nil
					}); err != nil {
						t.Fatal(err)
					}
					if mode == "id-configured" {
						args[2] = "task-one"
						want = 2
					}
				}
				var err error
				if kind == "thread" {
					err = runThread(h.io, args)
				} else {
					err = runAgents(h.io, args)
				}
				if exitCodeOf(t, err) != 0 || calls.Load() != want {
					t.Fatalf("scope/ID regression calls=%d: %s", calls.Load(), h.stdout.String())
				}
			})
		}
	}
}

func TestUserCPWorkKeyInvalidInputDoesNotOpenLogin(t *testing.T) {
	for _, kind := range []string{"thread", "agent"} {
		cases := [][]string{
			{"list", "--work", "SC-1"},
			{"list", "--work", "SC-1", "--team-id", ""},
			{"list", "--work", "SC-1", "--workspace", ""},
			{"list", "--work", "task-one"},
			{"list", "--work", "task-one", "--workspace", "workspace-one", "--team-id", "team-one"},
			{"list", "--work", "SC-1", "--workspace", "workspace-one", "--team", "team-one", "--team-id", "other-team"},
			{"show", "bad/id", "--work", "SC-1", "--workspace", "workspace-one"},
		}
		if kind == "agent" {
			cases = append(cases, []string{"list", "--work", "SC-1", "--workspace", "workspace-one", "--cursor", "01"})
		}
		for i, args := range cases {
			t.Run(fmt.Sprintf("%s/%d", kind, i), func(t *testing.T) {
				h := newContextHarness(t, nil)
				h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
					t.Fatal("invalid input opened credentials")
					return nil, nil
				}
				args = append(args, "--json")
				var err error
				if kind == "thread" {
					err = runThread(h.io, args)
				} else {
					err = runAgents(h.io, args)
				}
				if exitCodeOf(t, err) != 2 || len(h.daemon.requests) != 0 {
					t.Fatal("bad input accepted", err)
				}
			})
		}
	}
}

func TestUserCPWorkKeyAmbiguityRecovery(t *testing.T) {
	for _, kind := range []string{"thread", "agent"} {
		t.Run(kind, func(t *testing.T) {
			service, calls := cpWorkKeyFixture(t, kind, "conflict")
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			args := []string{"list", "--work", "SC-1", "--workspace", "workspace-one", "--json"}
			var err error
			if kind == "thread" {
				err = runThread(h.io, args)
			} else {
				err = runAgents(h.io, args)
			}
			env := decodeEnvelope(t, h.stdout.Bytes())
			if exitCodeOf(t, err) != int(contextcli.ExitConflict) || env.Error == nil || env.Error.Code != contextcli.CodeConflict || env.Data != nil || calls.Load() != 1 {
				t.Fatal("ambiguity not isolated", err, calls.Load())
			}
			if !strings.Contains(env.Error.NextAction, "--work <ID> --workspace <WORKSPACE-ID>") || strings.Contains(h.stdout.String(), "private-reference-canary") {
				t.Fatal("recovery must support ID/workspace", env.Error.NextAction)
			}
		})
	}
}
