package userauth

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestAgentReadOversizeNeverPublishesPartialBody(t *testing.T) {
	m, _ := threadReadTestManager(t, agentReadFixture, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, strings.Repeat("x", (16<<20)+1))
	})
	err := m.ReadTaskAgents(context.Background(), "workspace", "task", "", "", func([]byte, int) error {
		t.Error("oversize reached parser")
		return nil
	})
	if !errors.Is(err, ErrAgentReadIncomplete) {
		t.Fatal("oversized Agent response did not preserve incomplete error", err)
	}
}

func TestAgentReadRedirectNeverChoosesTarget(t *testing.T) {
	for _, at := range []string{"exchange", "CP"} {
		t.Run(at, func(t *testing.T) {
			calls := 0
			m, _ := threadReadTestManager(t, func(w http.ResponseWriter, r *http.Request) {
				if at == "exchange" {
					http.Redirect(w, r, "https://attacker.invalid/private-canary", http.StatusTemporaryRedirect)
					return
				}
				agentReadFixture(w, r)
			}, func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "https://attacker.invalid/private-canary", http.StatusTemporaryRedirect)
			})
			err := m.ReadTaskAgents(context.Background(), "workspace", "task", "", "", func(raw []byte, status int) error {
				calls++
				if raw != nil || status != http.StatusTemporaryRedirect {
					t.Error("redirect body or target consumed")
				}
				return errors.New("safe redirect refusal")
			})
			if err == nil || calls != 1 || strings.Contains(err.Error(), "private-canary") {
				t.Fatal("redirect was not safely refused", err)
			}
		})
	}
}
