package saasplane

import (
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// Text deltas must NOT be forwarded to the control plane. Providers (esp. codex)
// emit them as tiny token/JSON fragments; surfacing each as its own progress
// line produced fragmented, incoherent output. Structured progress + the final
// result are what reach the client.
func TestEventRequestFromAgentEventDropsTextDelta(t *testing.T) {
	assignment := assignmentcontract.Assignment{ID: "asn-1", TaskID: "task-1"}
	if _, ok := eventRequestFromAgentEvent(assignment, agentbridge.Event{
		Kind: agentbridge.EventTextDelta,
		Text: "hello world",
	}); ok {
		t.Fatal("text delta must not be forwarded as a progress event")
	}

	// Structured progress is still forwarded as a RiidoLog line.
	req, ok := eventRequestFromAgentEvent(assignment, agentbridge.Event{
		Kind: agentbridge.EventProgress,
		Text: "작업을 시작했어요.",
	})
	if !ok || req.EventType != assignmentcontract.EventRiidoLog {
		t.Fatalf("structured progress should still be forwarded: ok=%v req=%+v", ok, req)
	}
}
