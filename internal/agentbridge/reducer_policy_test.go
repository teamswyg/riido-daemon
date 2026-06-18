package agentbridge

import (
	"strings"
	"testing"
)

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
