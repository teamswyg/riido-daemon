package agentbridge

import (
	"strings"
	"testing"
)

func TestReduceSessionIdentifiedEmitsPersist(t *testing.T) {
	s := NewState()
	s, cmds := Reduce(s, Event{Kind: EventSessionIdentified, SessionID: "sess-1"}, nil)
	if s.SessionID != "sess-1" {
		t.Fatalf("session id not set: %q", s.SessionID)
	}
	if len(cmds) != 1 || cmds[0].Kind != CommandPersistSession {
		t.Fatalf("expected one CommandPersistSession, got %+v", cmds)
	}
}

func TestReduceTextDeltaAccumulatesOutput(t *testing.T) {
	s := NewState()
	s, _ = Reduce(s, Event{Kind: EventTextDelta, Text: "hello "}, nil)
	s, _ = Reduce(s, Event{Kind: EventTextDelta, Text: "world"}, nil)
	if string(s.Output) != "hello world" {
		t.Fatalf("output not accumulated: %q", s.Output)
	}
}

func TestReduceResultUsesAccumulatedOutputWhenResultOutputEmpty(t *testing.T) {
	s := NewState()
	s, _ = Reduce(s, Event{Kind: EventTextDelta, Text: "hello "}, nil)
	s, _ = Reduce(s, Event{Kind: EventTextDelta, Text: "world"}, nil)
	s, _ = Reduce(s, Event{Kind: EventResult, Result: Result{Status: ResultCompleted}}, nil)
	if s.Result.Output != "hello world" {
		t.Fatalf("result output: %q", s.Result.Output)
	}
}

// Invariant 1: terminal entered exactly once.
func TestReduceTerminalExactlyOnce(t *testing.T) {
	s := NewState()
	s, _ = Reduce(s, Event{Kind: EventResult, Result: Result{Status: ResultCompleted, Output: "done"}}, nil)
	if !s.Terminal {
		t.Fatalf("expected terminal after EventResult")
	}
	before := s
	s, _ = Reduce(s, Event{Kind: EventResult, Result: Result{Status: ResultFailed}}, nil)
	if s.Phase != before.Phase || s.Result.Status != before.Result.Status {
		t.Fatalf("terminal mutated by second EventResult: phase=%s status=%s", s.Phase, s.Result.Status)
	}
}

// Invariant 2: non-terminal events after terminal don't change phase,
// but usage may still accumulate per policy.
func TestReduceUsageAccumulatesAfterTerminal(t *testing.T) {
	s := NewState()
	s, _ = Reduce(s, Event{Kind: EventUsageDelta, Usage: Usage{PromptTokens: 10}}, nil)
	s, _ = Reduce(s, Event{Kind: EventResult, Result: Result{Status: ResultCompleted}}, nil)
	phaseBefore := s.Phase
	s, _ = Reduce(s, Event{Kind: EventUsageDelta, Usage: Usage{CompletionTokens: 5}}, nil)
	if s.Phase != phaseBefore {
		t.Fatalf("phase changed after terminal: %s -> %s", phaseBefore, s.Phase)
	}
	if s.Usage.PromptTokens != 10 || s.Usage.CompletionTokens != 5 {
		t.Fatalf("usage not accumulated across terminal: %+v", s.Usage)
	}
}

// Invariant 3: SessionIdentified accepted any time before terminal.
func TestReduceSessionIdentifiedLate(t *testing.T) {
	s := NewState()
	s, _ = Reduce(s, Event{Kind: EventTextDelta, Text: "hi"}, nil)
	s, _ = Reduce(s, Event{Kind: EventSessionIdentified, SessionID: "late"}, nil)
	if s.SessionID != "late" {
		t.Fatalf("session id not updated: %q", s.SessionID)
	}
}

// Invariant 4: Completed valid with empty Output.
func TestReduceCompletedEmpty(t *testing.T) {
	s := NewState()
	s, _ = Reduce(s, Event{Kind: EventResult, Result: Result{Status: ResultCompleted}}, nil)
	if !s.Terminal || s.Result.Status != ResultCompleted {
		t.Fatalf("expected completed terminal, got %s/%s", s.Phase, s.Result.Status)
	}
}

// Invariant 5: process exit 0 with provider failed result -> failed.
func TestReduceProviderFailedSurvivesZeroExit(t *testing.T) {
	s := NewState()
	s, _ = Reduce(s, Event{Kind: EventResult, Result: Result{Status: ResultFailed, Error: "boom"}}, nil)
	s, _ = Reduce(s, Event{Kind: EventProcessExit, ExitCode: 0}, nil)
	if s.Result.Status != ResultFailed {
		t.Fatalf("expected failed (provider-result wins), got %s", s.Result.Status)
	}
}

// Invariant 5b: process non-zero with no provider result -> failed.
func TestReduceProcessExitNonZeroNoResult(t *testing.T) {
	s := NewState()
	s, _ = Reduce(s, Event{Kind: EventProcessExit, ExitCode: 137, Err: "killed"}, nil)
	if !s.Terminal || s.Result.Status != ResultFailed {
		t.Fatalf("expected failed terminal, got %s/%s", s.Phase, s.Result.Status)
	}
}

// Invariant 6: cancellation preempts later events.
func TestReduceCancellationWinsOverLaterResult(t *testing.T) {
	s := NewState()
	s, cmds := Reduce(s, Event{Kind: EventCancellation, Err: "user"}, nil)
	if !s.Terminal || s.Result.Status != ResultCancelled {
		t.Fatalf("expected cancelled, got %s/%s", s.Phase, s.Result.Status)
	}
	if len(cmds) != 1 || cmds[0].Kind != CommandCancelProvider {
		t.Fatalf("expected CommandCancelProvider, got %+v", cmds)
	}
	s, _ = Reduce(s, Event{Kind: EventResult, Result: Result{Status: ResultCompleted}}, nil)
	if s.Result.Status != ResultCancelled {
		t.Fatalf("cancellation lost to later result: %s", s.Result.Status)
	}
}

// Invariant 7: timeout preempts later result.
func TestReduceTimeoutWins(t *testing.T) {
	s := NewState()
	s, _ = Reduce(s, Event{Kind: EventTimeout, Err: "deadline"}, nil)
	if s.Result.Status != ResultTimeout {
		t.Fatalf("expected timeout, got %s", s.Result.Status)
	}
	s, _ = Reduce(s, Event{Kind: EventResult, Result: Result{Status: ResultCompleted}}, nil)
	if s.Result.Status != ResultTimeout {
		t.Fatalf("timeout lost to later result: %s", s.Result.Status)
	}
}

// Invariant 8: no provider name in RunState.
func TestStatesAreProviderNeutral(t *testing.T) {
	providerNames := []string{
		"claude", "codex", "openclaw", "cursor",
		"anthropic", "openai", "gemini", "copilot",
	}
	for _, s := range AllStates() {
		lower := strings.ToLower(string(s))
		for _, name := range providerNames {
			if strings.Contains(lower, name) {
				t.Fatalf("state %q contains provider name %q", s, name)
			}
		}
	}
}

// Invariant 9: usage accumulates regardless of provider schema.
func TestReduceUsageMerges(t *testing.T) {
	s := NewState()
	s, _ = Reduce(s, Event{Kind: EventUsageDelta, Usage: Usage{PromptTokens: 1, CompletionTokens: 2}}, nil)
	s, _ = Reduce(s, Event{Kind: EventUsageDelta, Usage: Usage{PromptTokens: 3, ReasoningTokens: 4}}, nil)
	want := Usage{PromptTokens: 4, CompletionTokens: 2, ReasoningTokens: 4}
	if s.Usage != want {
		t.Fatalf("usage merge wrong: got %+v want %+v", s.Usage, want)
	}
}

// Invariant 10 (default policy): nil AutoApprover leaves the run in
// WaitingToolApproval. Auto-bypass is not the default.
func TestReduceToolApprovalDefaultRequiresHuman(t *testing.T) {
	s := NewState()
	s, cmds := Reduce(s, Event{Kind: EventToolApprovalNeeded, Tool: ToolRef{ID: "t1"}}, nil)
	if s.Phase != StateWaitingToolApproval {
		t.Fatalf("expected WaitingToolApproval, got %s", s.Phase)
	}
	if len(cmds) != 0 {
		t.Fatalf("default policy must not auto-approve, got cmds=%+v", cmds)
	}
}
