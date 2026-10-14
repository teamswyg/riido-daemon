package saasplane

import (
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestReportEventCoalescesTextDeltaIntoEvolvingBody(t *testing.T) {
	fake, plane, req := newTextDeltaFixture(t)
	reportTextDeltas(t, plane, req.ID, "Hello, ", "World! This is a streamed answer body.")

	last := lastTextDeltaEvent(t, fake)
	if last.EventType != assignmentcontract.EventRiidoLog {
		t.Fatalf("partial body should be a RiidoLog event, got %q", last.EventType)
	}
	if last.Message != "Hello, World! This is a streamed answer body." {
		t.Fatalf("partial body should carry the full accumulated text, got %q", last.Message)
	}
	if last.Metadata[agentbridge.ProgressMessageMetadataCode] != "9001" ||
		last.Metadata[agentbridge.ProgressMessageMetadataKey] != "assistant.partial" {
		t.Fatalf("partial body must be tagged with sentinel metadata = %+v", last.Metadata)
	}
}

func lastTextDeltaEvent(t *testing.T, fake *fakeAssignmentServer) assignmentcontract.AgentEventRequest {
	t.Helper()
	if len(fake.events) == 0 {
		t.Fatal("expected at least one forwarded partial-body event")
	}
	return fake.events[len(fake.events)-1]
}
