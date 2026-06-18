package saasplane

import (
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestEventRequestFromAgentEventDropsTextDelta(t *testing.T) {
	assignment := assignmentcontract.Assignment{ID: "asn-1", TaskID: "task-1"}
	if _, ok := eventRequestFromAgentEvent(assignment, agentbridge.Event{
		Kind: agentbridge.EventTextDelta,
		Text: "hello world",
	}); ok {
		t.Fatal("text delta must not be forwarded as a per-delta progress event")
	}

	req, ok := eventRequestFromAgentEvent(assignment, agentbridge.Event{
		Kind: agentbridge.EventProgress,
		Text: "작업을 시작했어요.",
	})
	if !ok || req.EventType != assignmentcontract.EventRiidoLog {
		t.Fatalf("structured progress should still be forwarded: ok=%v req=%+v", ok, req)
	}
}
