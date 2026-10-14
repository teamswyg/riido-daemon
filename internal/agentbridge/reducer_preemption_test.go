package agentbridge

import "testing"

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
