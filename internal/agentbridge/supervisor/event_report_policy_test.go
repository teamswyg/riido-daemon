package supervisor

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestShouldRetainStateAndWarningReportEvents(t *testing.T) {
	want := map[agentbridge.EventKind]bool{
		agentbridge.EventLifecycle:          true,
		agentbridge.EventSessionIdentified:  true,
		agentbridge.EventToolApprovalNeeded: true,
		agentbridge.EventWarning:            true,
		agentbridge.EventCancellation:       true,
		agentbridge.EventTimeout:            true,
		agentbridge.EventProcessExit:        true,
	}
	for _, kind := range agentbridge.EventKinds() {
		got := shouldRetainReportEvent(agentbridge.Event{Kind: kind})
		if got != want[kind] {
			t.Fatalf("retain %s = %v, want %v", kind, got, want[kind])
		}
	}
}
