package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

type threadCLITransport func(*http.Request) (*http.Response, error)

func (f threadCLITransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func threadCLIService(t *testing.T, body string, status int) (*standalonecli.Service, *atomic.Int32) {
	t.Helper()
	var calls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth/cli/thread-read-tickets" {
			if r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer fixture-access" {
				t.Error("exchange original OAuth required")
			}
			var q map[string]string
			if json.NewDecoder(r.Body).Decode(&q) != nil || len(q) != 2 || q["workspace_id"] != "workspace-one" || q["task_id"] != "task-one" {
				t.Error("wrong ticket scope")
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"schema_version": "riido-cli-thread-read.v1", "access_token": "riido_cli_thread_v1_" + strings.Repeat("A", 43), "token_type": "Bearer", "resource": "https://staging.ai-api.riido.io", "purpose": "task_threads_read", "expires_at": time.Now().Add(time.Minute).UTC().Format(time.RFC3339Nano)})
			return
		}
		if r.Method != http.MethodGet || r.URL.Path != "/v3/client/workspaces/workspace-one/ai-agent/tasks/task-one/threads" || r.Header.Get("Authorization") != "Bearer riido_cli_thread_v1_"+strings.Repeat("A", 43) {
			t.Error("CP received wrong credential or path")
		}
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(server.Close)
	client := &http.Client{Transport: threadCLITransport(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host != "monolith.api.riido.io" && r.URL.Host != "staging.ai-api.riido.io" {
			t.Fatal("unexpected target")
		}
		out := r.Clone(r.Context())
		u := *r.URL
		target, _ := url.Parse(server.URL)
		u.Scheme, u.Host = target.Scheme, target.Host
		out.URL, out.Host = &u, target.Host
		return server.Client().Transport.RoundTrip(out)
	})}
	store := &cliMemoryStore{present: true, credential: userauth.Credential{SessionID: "fixture-session", AccessToken: "fixture-access", RefreshToken: "fixture-refresh", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour)}}
	service, err := standalonecli.New(standalonecli.Config{APIOrigin: "https://monolith.api.riido.io", Store: store, HTTPClient: client})
	if err != nil {
		t.Fatal(err)
	}
	return service, &calls
}

const threadCLIHistory = `{"schema_version":"riido-ai-server.v1","task_id":"task-one","assigned_agent_ids":["agent-one"],"threads":[{"thread_id":"thread-one","conversation_id":"conversation-one","task_id":"task-one","agent_id":"agent-one","run_id":"run-one","work_status":"done","assignment_state":"completed","messages":[{"message_id":"message-one","role":"agent","body":"답변\u001b[31m"}]}]}`

func TestThreadCLIHTTPListAndShow(t *testing.T) {
	for _, command := range []string{"list", "show"} {
		t.Run(command, func(t *testing.T) {
			s, calls := threadCLIService(t, threadCLIHistory, 200)
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return s, nil }
			args := []string{command}
			if command == "show" {
				args = append(args, "thread-one")
			}
			args = append(args, "--work", "task-one", "--workspace", "workspace-one", "--json")
			if err := runThread(h.io, args); err != nil {
				t.Fatal(err)
			}
			env := decodeEnvelope(t, h.stdout.Bytes())
			data := env.Data.(map[string]any)
			history := data["history"].(map[string]any)
			if history["complete"] != false || history["projection_complete"] != true || history["history_scope"] != "server_retained" || len(env.Warnings) != 1 || calls.Load() != 2 {
				t.Fatal("CLI lost bounded-history evidence")
			}
			for _, secret := range []string{"fixture-access", "fixture-refresh", "riido_cli_thread_v1_"} {
				if strings.Contains(h.stdout.String()+h.stderr.String(), secret) {
					t.Fatal("credential exposed")
				}
			}
			if command == "show" {
				thread := history["thread"].(map[string]any)
				if thread["conversation_id"] != "conversation-one" || thread["run_id"] != "run-one" || len(thread["messages"].([]any)) != 1 {
					t.Fatal("thread identity lost")
				}
			}
		})
	}
}

func TestThreadCLIFailsBeforeOpeningCredentials(t *testing.T) {
	for _, args := range [][]string{
		{"list"},
		{"show", "--work", "task-one", "--workspace", "workspace-one"},
		{"list", "extra", "--work", "task-one", "--workspace", "workspace-one"},
		{"list", "--work", "../task", "--workspace", "workspace-one"},
		{"list", "--work", "RIID-5972", "--workspace", "workspace-one"},
		{"list", "--work", "task-one", "--workspace", "workspace-one", "--cursor", "x"},
		{"list", "--work", "task-one", "--workspace", "workspace-one", "--limit", "5"},
		{"list", "--work", "task-one", "--workspace", "workspace-one", "--page", "1"},
		{"list", "--work", "task-one", "--workspace", "workspace-one", "--budget", "1"},
		{"list", "--token", "secret"},
		{"list", "--cp-url", "https://attacker.invalid"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
				t.Error("opened protected credentials")
				return nil, nil
			}
			if err := runThread(h.io, append(args, "--json")); err == nil {
				t.Fatal("invalid arguments accepted")
			}
		})
	}
	h := newContextHarness(t, map[string]string{"RIIDO_RUN_TOKEN": "fixture"})
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
		t.Error("agent opened user session")
		return nil, nil
	}
	if err := runThread(h.io, []string{"list", "--work", "task-one", "--workspace", "workspace-one", "--json"}); err == nil {
		t.Fatal("agent mode accepted")
	}
}

func TestThreadCLIErrorAndTerminalSafety(t *testing.T) {
	for _, tc := range []struct {
		name, body, id string
		status         int
	}{{"malformed", "{", "thread-one", 200}, {"denied", "secret error", "thread-one", 403}, {"missing", threadCLIHistory, "missing", 200}} {
		t.Run(tc.name, func(t *testing.T) {
			s, _ := threadCLIService(t, tc.body, tc.status)
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return s, nil }
			if err := runThread(h.io, []string{"show", tc.id, "--work", "task-one", "--workspace", "workspace-one", "--json"}); err == nil {
				t.Fatal("failed read reported success")
			}
			env := decodeEnvelope(t, h.stdout.Bytes())
			if env.Error == nil || env.Data != nil || strings.Contains(h.stdout.String(), "secret error") {
				t.Fatal("error leaked history or provider body")
			}
		})
	}
	s, _ := threadCLIService(t, threadCLIHistory, 200)
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return s, nil }
	if err := runThread(h.io, []string{"show", "thread-one", "--work", "task-one", "--workspace", "workspace-one"}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(h.stdout.String(), "\x1b") || !strings.Contains(h.stdout.String(), "\\x1b") {
		t.Fatal("terminal escape not quoted")
	}
}
