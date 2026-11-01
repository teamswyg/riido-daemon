package userauth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestAgentReadRejectsCrossPurposeAndMalformedTicket(t *testing.T) {
	for _, kind := range []string{"thread", "schema", "purpose", "origin", "OAuth", "expired", "long", "null", "extra", "duplicate", "alias", "trailing", "noncanonical"} {
		t.Run(kind, func(t *testing.T) {
			var reads atomic.Int32
			m, _ := threadReadTestManager(t, func(w http.ResponseWriter, _ *http.Request) {
				v := agentReadTicketFixture()
				switch kind {
				case "thread":
					v = threadReadTicketFixture()
				case "schema":
					v["schema_version"] = "riido-cli-thread-read.v1"
				case "purpose":
					v["purpose"] = "task_threads_read"
				case "origin":
					v["resource"] = "https://other.invalid"
				case "OAuth":
					v["access_token"] = "old-access"
				case "expired":
					v["expires_at"] = time.Now().Add(-time.Minute).Format(time.RFC3339)
				case "long":
					v["expires_at"] = time.Now().Add(time.Hour).Format(time.RFC3339)
				case "null":
					v["purpose"] = nil
				case "extra":
					v["url"] = "private-canary"
				case "alias":
					v["Purpose"] = v["purpose"]
					delete(v, "purpose")
				case "noncanonical":
					v["access_token"] = "riido_cli_agent_v1_" + strings.Repeat("A", 42) + "B"
				}
				body, _ := json.Marshal(v)
				if kind == "duplicate" {
					body = append([]byte(`{"purpose":"task_agents_read",`), body[1:]...)
				}
				if kind == "trailing" {
					body = append(body, []byte(` {}`)...)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(body)
			}, func(http.ResponseWriter, *http.Request) { reads.Add(1) })
			err := m.ReadTaskAgents(context.Background(), "workspace", "task", "", "", func([]byte, int) error { t.Error("invalid ticket reached parser"); return nil })
			if err == nil || reads.Load() != 0 || strings.Contains(err.Error(), "private-canary") {
				t.Fatal("invalid ticket reached CP or leaked", err)
			}
		})
	}
}

func TestAgentReadRejectsInputBeforeProtectedStorage(t *testing.T) {
	for _, tc := range []struct{ origin, workspace, task, agent, cursor string }{
		{"https://other.invalid", "w", "t", "", ""}, {"", "../w", "t", "", ""}, {"", "w", "t/x", "", ""}, {"", "w", "t", "a?x", ""}, {"", "w", "t", "", "01"}, {"", "w", "t", "", "-1"}, {"", "w", "t", "", "1&cursor=2"}, {"", "w", "t", "a", "0"},
	} {
		m, store := threadReadTestManager(t, func(http.ResponseWriter, *http.Request) { t.Error("unexpected exchange") }, func(http.ResponseWriter, *http.Request) { t.Error("unexpected CP") })
		if tc.origin != "" {
			m.client.config.Resource = tc.origin
		}
		if err := m.ReadTaskAgents(context.Background(), tc.workspace, tc.task, tc.agent, tc.cursor, func([]byte, int) error { return nil }); err == nil || store.loads != 0 {
			t.Fatal("invalid input reached protected storage", tc, err, store.loads)
		}
	}
}
