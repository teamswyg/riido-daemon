package agentbridge

import "testing"

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

func TestReduceProviderFailedSurvivesZeroExit(t *testing.T) {
	s := NewState()
	s, _ = Reduce(s, Event{Kind: EventResult, Result: Result{Status: ResultFailed, Error: "boom"}}, nil)
	s, _ = Reduce(s, Event{Kind: EventProcessExit, ExitCode: 0}, nil)
	if s.Result.Status != ResultFailed {
		t.Fatalf("expected failed (provider-result wins), got %s", s.Result.Status)
	}
}

func TestReduceProcessExitNonZeroNoResult(t *testing.T) {
	s := NewState()
	s, _ = Reduce(s, Event{Kind: EventProcessExit, ExitCode: 137, Err: "killed"}, nil)
	if !s.Terminal || s.Result.Status != ResultFailed {
		t.Fatalf("expected failed terminal, got %s/%s", s.Phase, s.Result.Status)
	}
}
