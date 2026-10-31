package userauth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type threadReadTestTransport func(*http.Request) (*http.Response, error)

func (f threadReadTestTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func threadReadTestManager(t *testing.T, mono, cp http.HandlerFunc) (*Manager, *fakeStore) {
	t.Helper()
	a, b := httptest.NewTLSServer(mono), httptest.NewTLSServer(cp)
	t.Cleanup(a.Close)
	t.Cleanup(b.Close)
	client := &http.Client{Transport: threadReadTestTransport(func(r *http.Request) (*http.Response, error) {
		s := a
		if r.URL.Host == "staging.ai-api.riido.io" {
			s = b
		} else if r.URL.Host != "monolith.api.riido.io" {
			t.Fatal("untrusted host reached transport")
		}
		out := r.Clone(r.Context())
		u := *r.URL
		target, _ := url.Parse(s.URL)
		u.Scheme, u.Host = target.Scheme, target.Host
		out.URL, out.Host = &u, target.Host
		return s.Client().Transport.RoundTrip(out)
	})}
	c, err := NewClient(Config{ClientID: "riido-cli", Scopes: []string{"read"}, Resource: "https://monolith.api.riido.io", HTTPClient: client, Endpoints: Endpoints{AuthorizeURL: "https://monolith.api.riido.io/oauth/authorize", TokenURL: "https://monolith.api.riido.io/oauth/token", RevokeURL: "https://monolith.api.riido.io/oauth/revoke"}})
	if err != nil {
		t.Fatal(err)
	}
	store := newFakeStore(validCredential())
	return NewManager(c, store), store
}

func threadReadTicketFixture() map[string]any {
	return map[string]any{"schema_version": "riido-cli-thread-read.v1", "access_token": "riido_cli_thread_v1_" + strings.Repeat("A", 43), "token_type": "Bearer", "resource": "https://staging.ai-api.riido.io", "purpose": "task_threads_read", "expires_at": time.Now().Add(time.Minute).UTC().Format(time.RFC3339Nano)}
}

func TestThreadReadSeparatesOAuthAndTicketOverActualHTTP(t *testing.T) {
	var exchanges, reads atomic.Int32
	m, store := threadReadTestManager(t, func(w http.ResponseWriter, r *http.Request) {
		exchanges.Add(1)
		var request map[string]string
		if r.Method != http.MethodPost || r.URL.Path != "/oauth/cli/thread-read-tickets" || r.Header.Get("Authorization") != "Bearer old-access" || r.Header.Get("Cookie") != "" || json.NewDecoder(r.Body).Decode(&request) != nil || len(request) != 2 || request["workspace_id"] != "workspace-one" || request["task_id"] != "task-one" {
			t.Error("original OAuth exchange changed")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(threadReadTicketFixture())
	}, func(w http.ResponseWriter, r *http.Request) {
		reads.Add(1)
		if r.Method != http.MethodGet || r.URL.Path != "/v3/client/workspaces/workspace-one/ai-agent/tasks/task-one/threads" || r.Header.Get("Authorization") != "Bearer riido_cli_thread_v1_"+strings.Repeat("A", 43) || r.Header.Get("Cookie") != "" {
			t.Error("original OAuth or wrong scope reached CP")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"retained":"fixture"}`)
	})
	for range 2 {
		raw, status, err := m.ReadTaskThreads(context.Background(), "workspace-one", "task-one")
		if err != nil || status != http.StatusOK || string(raw) != `{"retained":"fixture"}` {
			t.Fatal("two-hop read failed", status, err)
		}
	}
	if exchanges.Load() != 2 || reads.Load() != 2 || store.loads != 2 {
		t.Fatal("ticket cached or protected access repeated", exchanges.Load(), reads.Load(), store.loads)
	}
}
