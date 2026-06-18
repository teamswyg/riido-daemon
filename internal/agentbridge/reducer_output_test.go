package agentbridge

import "testing"

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

func TestReduceCompletedEmpty(t *testing.T) {
	s := NewState()
	s, _ = Reduce(s, Event{Kind: EventResult, Result: Result{Status: ResultCompleted}}, nil)
	if !s.Terminal || s.Result.Status != ResultCompleted {
		t.Fatalf("expected completed terminal, got %s/%s", s.Phase, s.Result.Status)
	}
}
