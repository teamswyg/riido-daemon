package agentbridge

import (
	"testing"
	"time"
)

func TestReduceToolApprovalExplicitApprover(t *testing.T) {
	s := NewState()
	approver := AutoApprover(func(tool ToolRef) bool { return tool.Kind == "read" })
	_, cmds := Reduce(s, Event{Kind: EventToolApprovalNeeded, Tool: ToolRef{ID: "t1", ProviderRequestID: "r1", Kind: "read"}}, approver)
	if len(cmds) != 1 || cmds[0].Kind != CommandApproveTool || cmds[0].ToolID != "t1" || cmds[0].ProviderRequestID != "r1" {
		t.Fatalf("expected approve t1, got %+v", cmds)
	}
}

// Semantic activity updates the watchdog timestamp; log spam does not.
func TestReduceSemanticActivityWatchdog(t *testing.T) {
	now := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	s := NewState()
	s, _ = Reduce(s, Event{Kind: EventLog, At: now}, nil)
	if !s.LastSemanticActivity.IsZero() {
		t.Fatalf("log should not count as semantic activity")
	}
	s, _ = Reduce(s, Event{Kind: EventTextDelta, At: now, Text: "x"}, nil)
	if !s.LastSemanticActivity.Equal(now) {
		t.Fatalf("text delta should update watchdog, got %v", s.LastSemanticActivity)
	}
	approvalAt := now.Add(time.Second)
	s, _ = Reduce(s, Event{Kind: EventToolApprovalNeeded, At: approvalAt, Tool: ToolRef{ID: "approval-1"}}, nil)
	if !s.LastSemanticActivity.Equal(approvalAt) {
		t.Fatalf("approval request should reset watchdog, got %v", s.LastSemanticActivity)
	}
}

// Tool start moves to ToolRunning; tool completion returns to Running
// when the in-flight set is empty.
func TestReduceToolCallLifecycle(t *testing.T) {
	s := NewState()
	s, _ = Reduce(s, Event{Kind: EventLifecycle, Phase: StateRunning}, nil)
	s, _ = Reduce(s, Event{Kind: EventToolCallStarted, Tool: ToolRef{ID: "t1"}}, nil)
	if s.Phase != StateToolRunning {
		t.Fatalf("expected ToolRunning, got %s", s.Phase)
	}
	s, _ = Reduce(s, Event{Kind: EventToolCallCompleted, Tool: ToolRef{ID: "t1"}}, nil)
	if s.Phase != StateRunning {
		t.Fatalf("expected Running after tool completion, got %s", s.Phase)
	}
}

// Lifecycle events must not be able to push the run into a terminal
// state — terminals only come from Result / Cancellation / Timeout /
// ProcessExit.
func TestReduceLifecycleCannotForceTerminal(t *testing.T) {
	s := NewState()
	s, _ = Reduce(s, Event{Kind: EventLifecycle, Phase: StateCompleted}, nil)
	if s.Terminal {
		t.Fatalf("lifecycle should not be able to mark terminal")
	}
	if s.Phase == StateCompleted {
		t.Fatalf("lifecycle should refuse to set terminal phase, got %s", s.Phase)
	}
}
