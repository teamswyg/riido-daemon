package agentbridge

import "testing"

func TestRunStateCodeRoundTrip(t *testing.T) {
	for _, state := range AllStates() {
		code := state.Code()
		if !code.IsKnown() {
			t.Fatalf("%s code is unknown", state)
		}
		if got := code.RunState(); got != state {
			t.Fatalf("%s code round trip = %s", state, got)
		}
	}
}

func TestRunStateTerminalUsesCodeLayer(t *testing.T) {
	terminal := map[RunState]bool{
		StateCompleted:   true,
		StateFailed:      true,
		StateCancelled:   true,
		StateTimedOut:    true,
		StateIdleStopped: true,
	}
	for _, state := range AllStates() {
		if got := state.IsTerminal(); got != terminal[state] {
			t.Fatalf("%s IsTerminal = %v, want %v", state, got, terminal[state])
		}
	}
}
