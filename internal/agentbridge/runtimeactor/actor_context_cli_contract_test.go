package runtimeactor

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestSubmitStartRequestContextCLIContractPlacement(t *testing.T) {
	for _, provider := range []string{"claude", "codex", "cursor", "openclaw"} {
		t.Run(provider, func(t *testing.T) {
			msg := &submitMsg{req: bridge.TaskRequest{ID: "task-1", Provider: bridge.Provider(provider), Prompt: "Keep my task.", SystemPrompt: "Existing policy."}}
			env := map[string]string{"RIIDO_RUN_TOKEN": "cap-secret", "RIIDO_DAEMON_SOCKET": "/tmp/agentd.sock", "RIIDO_TASK_ID": "task-1"}
			req := submitStartRequest(msg, Capability{}, "rt-1", env, nil, true, "")
			combined := req.Prompt + req.SystemPrompt
			for _, want := range []string{"riido context get --json --no-input", "assignment-time snapshot", "work comments", "returned cursor", "not supported", "Never print", "agents list --json", "ask a person"} {
				if !strings.Contains(combined, want) {
					t.Errorf("missing CLI rule %q", want)
				}
			}
			if strings.Contains(combined, "cap-secret") || strings.Contains(combined, "/tmp/agentd.sock") {
				t.Fatal("CLI wiring leaked into instructions")
			}
			if provider != "cursor" {
				if req.Prompt != msg.req.Prompt || !strings.Contains(req.SystemPrompt, "Riido context CLI") {
					t.Fatal("native system placement lost")
				}
			} else if !strings.HasPrefix(req.Prompt, msg.req.Prompt) || req.SystemPrompt != msg.req.SystemPrompt {
				t.Fatal("inline placement lost task-first order")
			}
			off := submitStartRequest(msg, Capability{}, "rt-1", env, nil, false, "")
			if off.Prompt != msg.req.Prompt || off.SystemPrompt != msg.req.SystemPrompt {
				t.Fatal("environment alone enabled CLI instructions without a launch hook")
			}
			for _, key := range []string{"RIIDO_RUN_TOKEN", "RIIDO_DAEMON_SOCKET", "RIIDO_TASK_ID"} {
				previous := env[key]
				delete(env, key)
				off := submitStartRequest(msg, Capability{}, "rt-1", env, nil, true, "")
				if off.Prompt != msg.req.Prompt || off.SystemPrompt != msg.req.SystemPrompt {
					t.Fatalf("incomplete wiring (%s) changed instructions", key)
				}
				env[key] = previous
			}
		})
	}
}
