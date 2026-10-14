package saasplane

import (
	"context"
	"net/http"
	"testing"

	"github.com/teamswyg/riido-contracts/metadatakeys"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestPlaneRetriesTransientEventPostWithIdempotencyKey(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.enqueue(queuedHTTPAssignment("asn-1", "hello"))
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	fake.failNext("/v1/agents/jykim1/events", 1, http.StatusServiceUnavailable)
	err = plane.ReportEvent(context.Background(), req.ID, agentbridge.Event{
		Kind: agentbridge.EventProgress,
		Text: "progress",
	})
	if err != nil {
		t.Fatalf("ReportEvent should retry transient event failure: %v", err)
	}
	if got := fake.requestCount("/v1/agents/jykim1/events"); got != 2 {
		t.Fatalf("event request count = %d, want 2", got)
	}
	if len(fake.events) != 1 {
		t.Fatalf("events = %+v, want one committed event", fake.events)
	}
	eventKey := metadatakeys.AssignmentEventKey.String()
	if got := fake.events[0].Metadata[eventKey]; got == "" {
		t.Fatalf("assignment event key metadata missing: %+v", fake.events[0].Metadata)
	}
}
