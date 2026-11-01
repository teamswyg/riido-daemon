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

	"github.com/teamswyg/riido-daemon/internal/standalonecli"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func userAgentsCLIService(t *testing.T, body string, status int) (*standalonecli.Service, *atomic.Int32) {
	t.Helper()
	var calls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth/cli/agent-read-tickets" {
			if r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer fixture-access" {
				t.Error("exchange requires original OAuth")
			}
			var q map[string]string
			if json.NewDecoder(r.Body).Decode(&q) != nil || len(q) != 2 || q["workspace_id"] != "workspace-one" || q["task_id"] != "task-one" {
				t.Error("incorrect ticket scope")
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"schema_version": "riido-cli-agent-read.v1", "access_token": "riido_cli_agent_v1_" + strings.Repeat("A", 43), "token_type": "Bearer", "resource": "https://staging.ai-api.riido.io", "purpose": "task_agents_read", "expires_at": time.Now().Add(time.Minute).UTC().Format(time.RFC3339Nano)})
			return
		}
		if r.Method != http.MethodGet || !strings.HasPrefix(r.URL.Path, "/v1/cli/workspaces/workspace-one/tasks/task-one/agents") || r.Header.Get("Authorization") != "Bearer riido_cli_agent_v1_"+strings.Repeat("A", 43) {
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
