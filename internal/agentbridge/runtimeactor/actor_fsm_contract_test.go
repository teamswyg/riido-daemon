package runtimeactor

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestRuntimeActorDoesNotCreateProviderSpecificFSM(t *testing.T) {
	for _, s := range agentbridge.AllStates() {
		lower := strings.ToLower(string(s))
		for _, p := range []string{"claude", "codex", "openclaw", "cursor"} {
			if strings.Contains(lower, p) {
				t.Fatalf("agentbridge RunState %q leaked provider name", s)
			}
		}
	}
}
