package session

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSessionToolApprovalGateBlocksHeadlessApproval(t *testing.T) {
	adapter := &recordingAdapter{
		name:   "fake",
		parser: &recordingParser{},
		translateFn: func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
			if string(raw.Bytes) != "APPROVAL" {
				return nil, nil, nil
			}
			return []agentbridge.Event{approvalNeededEvent()}, nil, nil
		},
	}
	scenario := startToolGateScenario(t, "task-tool-approval-block", adapter, func(cfg *Config) {
		cfg.ToolApprovalGate = func(tool agentbridge.ToolRef) agentbridge.ToolStartDecision {
			if tool.ID != "approval-1" {
				t.Fatalf("unexpected tool: %+v", tool)
			}
			return agentbridge.ToolStartDecision{
				Block:  true,
				Code:   "TOOL_USE_NOT_IN_POLICY_BUNDLE",
				Reason: "no headless approval path",
			}
		}
	})

	scenario.running.EmitStdout([]byte("APPROVAL"))
	expectToolProviderKill(t, scenario.running)
	res := waitResult(t, scenario.session, 2*time.Second)
	if res.Status != agentbridge.ResultBlocked {
		t.Fatalf("result: %+v", res)
	}
	if res.Error != "TOOL_USE_NOT_IN_POLICY_BUNDLE: no headless approval path" {
		t.Fatalf("block error: %q", res.Error)
	}
	events := drainEvents(t, scenario.session, time.Second)
	if !hasWarningText(events, "tool approval unavailable in headless run") {
		t.Fatalf("missing headless approval warning in events: %+v", events)
	}
}

func approvalNeededEvent() agentbridge.Event {
	return agentbridge.Event{
		Kind: agentbridge.EventToolApprovalNeeded,
		Tool: agentbridge.ToolRef{ID: "approval-1", Kind: "patch_apply"},
	}
}
