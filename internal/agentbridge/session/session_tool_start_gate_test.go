package session

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSessionToolStartGateBlocksStartedTool(t *testing.T) {
	adapter := &recordingAdapter{
		name:   "fake",
		parser: &recordingParser{},
		translateFn: func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
			if string(raw.Bytes) != "START" {
				return nil, nil, nil
			}
			return []agentbridge.Event{startedShellToolEvent()}, nil, nil
		},
	}
	scenario := startToolGateScenario(t, "task-tool-start-block", adapter, func(cfg *Config) {
		cfg.ToolStartGate = func(tool agentbridge.ToolRef) agentbridge.ToolStartDecision {
			if tool.ID != "tool-1" {
				t.Fatalf("unexpected tool: %+v", tool)
			}
			return agentbridge.ToolStartDecision{
				Block:  true,
				Code:   "TOOL_USE_NOT_IN_POLICY_BUNDLE",
				Reason: "blocked in test",
			}
		}
	})

	scenario.running.EmitStdout([]byte("START"))
	expectToolProviderKill(t, scenario.running)
	res := waitResult(t, scenario.session, 2*time.Second)
	if res.Status != agentbridge.ResultBlocked {
		t.Fatalf("result: %+v", res)
	}
	if res.Error != "TOOL_USE_NOT_IN_POLICY_BUNDLE: blocked in test" {
		t.Fatalf("block error: %q", res.Error)
	}
	events := drainEvents(t, scenario.session, time.Second)
	if !hasWarningText(events, "tool use blocked by policy") {
		t.Fatalf("missing policy warning in events: %+v", events)
	}
}

func startedShellToolEvent() agentbridge.Event {
	return agentbridge.Event{
		Kind: agentbridge.EventToolCallStarted,
		Tool: agentbridge.ToolRef{
			ID:   "tool-1",
			Kind: "shell",
			Args: map[string]string{"command": "rm -rf .git"},
		},
	}
}
