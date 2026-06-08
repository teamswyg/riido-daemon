package saasplane

import (
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// Assistant text deltas must be forwarded to the control plane as RiidoLog
// progress so the web client's SSE thread stream shows the answer live. Empty
// deltas carry no content and must be dropped.
func TestEventRequestFromAgentEventForwardsTextDelta(t *testing.T) {
	assignment := assignmentcontract.Assignment{ID: "asn-1", TaskID: "task-1"}

	req, ok := eventRequestFromAgentEvent(assignment, agentbridge.Event{
		Kind: agentbridge.EventTextDelta,
		Text: "hello world",
	})
	if !ok {
		t.Fatal("text delta should be forwarded to the control plane")
	}
	if req.EventType != assignmentcontract.EventRiidoLog {
		t.Fatalf("text delta event type = %q, want %q", req.EventType, assignmentcontract.EventRiidoLog)
	}
	if req.Message != "hello world" {
		t.Fatalf("text delta message = %q, want %q", req.Message, "hello world")
	}
	if req.AssignmentID != "asn-1" || req.TaskID != "task-1" {
		t.Fatalf("text delta lost assignment/task: %+v", req)
	}

	if _, ok := eventRequestFromAgentEvent(assignment, agentbridge.Event{
		Kind: agentbridge.EventTextDelta,
		Text: "",
	}); ok {
		t.Fatal("empty text delta should be dropped, not forwarded")
	}
}
