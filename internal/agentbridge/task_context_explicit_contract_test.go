package agentbridge

import (
	"strings"
	"testing"
)

func TestContextCLIContractExplainsExplicitLiveReads(t *testing.T) {
	text := TaskContextCLIContractInstruction()
	for _, required := range []string{"work activity", "work body", "thread list", "thread show", "current permissions", "--version", "--page", "--comment-id", "live"} {
		if !strings.Contains(text, required) {
			t.Errorf("missing live read contract %q", required)
		}
	}
	for _, stale := range []string{"work activity and work search are not supported", "CLI data is an assignment-time snapshot, not a live read"} {
		if strings.Contains(text, stale) {
			t.Errorf("stale launch instruction: %s", stale)
		}
	}
}
