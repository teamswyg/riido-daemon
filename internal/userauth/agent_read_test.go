package userauth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func agentReadTicketFixture() map[string]any {
	v := threadReadTicketFixture()
	v["schema_version"], v["purpose"] = "riido-cli-agent-read.v1", "task_agents_read"
	v["access_token"] = "riido_cli_agent_v1_" + strings.Repeat("A", 43)
	return v
}

func agentReadFixture(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(agentReadTicketFixture())
}

func TestAgentReadActualHTTPAuthorityAndLockThroughParsing(t *testing.T) {
	var exchanges, reads atomic.Int32
	m, store := threadReadTestManager(t, func(w http.ResponseWriter, r *http.Request) {
		exchanges.Add(1)
		var q map[string]string
		if r.Method != http.MethodPost || r.URL.Path != "/oauth/cli/agent-read-tickets" || r.Header.Get("Authorization") != "Bearer old-access" || r.Header.Get("Cookie") != "" || json.NewDecoder(r.Body).Decode(&q) != nil || len(q) != 2 || q["workspace_id"] != "workspace-one" || q["task_id"] != "task-one" {
			t.Error("wrong original OAuth or task scope")
		}
		agentReadFixture(w, r)
	}, func(w http.ResponseWriter, r *http.Request) {
		reads.Add(1)
		if r.Method != http.MethodGet || r.URL.RequestURI() != "/v1/cli/workspaces/workspace-one/tasks/task-one/agents?cursor=1" || r.Header.Get("Authorization") != "Bearer riido_cli_agent_v1_"+strings.Repeat("A", 43) || r.Header.Get("Cookie") != "" {
			t.Error("wrong CP credential or exact page")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{}`)
	})
	other := NewManager(m.client, store)
	for range 2 {
		err := m.ReadTaskAgents(context.Background(), "workspace-one", "task-one", "", "1", func(raw []byte, status int) error {
			if string(raw) != `{}` || status != 200 {
				t.Fatal("response changed")
			}
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
			defer cancel()
			if _, err := other.AccessToken(ctx); !errors.Is(err, context.DeadlineExceeded) {
				t.Error("protected lock released before parsing", err)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if exchanges.Load() != 2 || reads.Load() != 2 {
		t.Fatal("ticket cached or extra HTTP calls")
	}
	if _, err := other.AccessToken(context.Background()); err != nil {
		t.Fatal("lock not released", err)
	}
}

func TestAgentReadDiscardsAccountSwitchAndCancellationAfterParsing(t *testing.T) {
	for _, kind := range []string{"session", "generation", "cancel"} {
		t.Run(kind, func(t *testing.T) {
			m, store := threadReadTestManager(t, agentReadFixture, func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, `{}`)
			})
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			err := m.ReadTaskAgents(ctx, "workspace", "task", "agent", "", func([]byte, int) error {
				switch kind {
				case "session":
					store.credential.SessionID = "replacement"
				case "generation":
					m.generation.Add(1)
				case "cancel":
					cancel()
				}
				return nil
			})
			if kind == "cancel" && !errors.Is(err, context.Canceled) || kind != "cancel" && !errors.Is(err, ErrLoggedOut) {
				t.Fatal("published stale parsed response", err)
			}
		})
	}
}
