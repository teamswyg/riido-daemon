package session

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSessionToolStartGateAllowsStartedTool(t *testing.T) {
	scenario := startToolGateScenario(t, "task-tool-start-allow", allowGateAdapter(), func(cfg *Config) {
		cfg.ToolStartGate = func(tool agentbridge.ToolRef) agentbridge.ToolStartDecision {
			if tool.ID != "tool-1" {
				t.Fatalf("unexpected tool: %+v", tool)
			}
			return agentbridge.ToolStartDecision{}
		}
	})

	scenario.running.EmitStdout([]byte("START"))
	emitDone(scenario.running)
	res := waitResult(t, scenario.session, 2*time.Second)
	if res.Status != agentbridge.ResultCompleted {
		t.Fatalf("result: %+v", res)
	}
	assertNoToolBlockWarning(t, scenario.session)
}

func TestSessionToolApprovalGateAllowsHeadlessApproval(t *testing.T) {
	scenario := startToolGateScenario(t, "task-tool-approval-allow", allowGateAdapter(), func(cfg *Config) {
		cfg.ToolApprovalGate = func(tool agentbridge.ToolRef) agentbridge.ToolStartDecision {
			if tool.ID != "approval-1" {
				t.Fatalf("unexpected tool: %+v", tool)
			}
			return agentbridge.ToolStartDecision{}
		}
	})

	scenario.running.EmitStdout([]byte("APPROVAL"))
	emitDone(scenario.running)
	res := waitResult(t, scenario.session, 2*time.Second)
	if res.Status != agentbridge.ResultCompleted {
		t.Fatalf("result: %+v", res)
	}
	assertNoToolBlockWarning(t, scenario.session)
}

func allowGateAdapter() *recordingAdapter {
	return &recordingAdapter{
		name:   "fake",
		parser: &recordingParser{},
		translateFn: func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
			switch string(raw.Bytes) {
			case "START":
				return []agentbridge.Event{startedShellToolEvent()}, nil, nil
			case "APPROVAL":
				return []agentbridge.Event{approvalNeededEvent()}, nil, nil
			case "DONE":
				return []agentbridge.Event{completedResultEvent("")}, nil, nil
			default:
				return nil, nil, nil
			}
		},
	}
}

func assertNoToolBlockWarning(t *testing.T, sess *Session) {
	t.Helper()
	events := drainEvents(t, sess, time.Second)
	if hasWarningText(events, "tool use blocked by policy") ||
		hasWarningText(events, "tool approval unavailable in headless run") {
		t.Fatalf("unexpected tool gate warning: %+v", events)
	}
}
