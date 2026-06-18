package agentbridge

import "testing"

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

func TestReduceUsageMerges(t *testing.T) {
	s := NewState()
	s, _ = Reduce(s, Event{Kind: EventUsageDelta, Usage: Usage{PromptTokens: 1, CompletionTokens: 2}}, nil)
	s, _ = Reduce(s, Event{Kind: EventUsageDelta, Usage: Usage{PromptTokens: 3, ReasoningTokens: 4}}, nil)
	want := Usage{PromptTokens: 4, CompletionTokens: 2, ReasoningTokens: 4}
	if s.Usage != want {
		t.Fatalf("usage merge wrong: got %+v want %+v", s.Usage, want)
	}
}
