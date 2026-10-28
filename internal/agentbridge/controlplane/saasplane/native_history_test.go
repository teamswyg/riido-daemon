package saasplane

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestNativeHistoryMappingPreservesSourceAndLease(t *testing.T) {
	m := agentbridge.NativeHistoryMessage{SessionID: "native", MessageID: "turn:item", Role: "agent", Body: "answer", Ordinal: 4}
	req, ok := eventRequestFromAgentEvent(assignment.Assignment{ID: "run", TaskID: "task", LeaseToken: "lease"}, agentbridge.Event{Kind: agentbridge.EventProgress, NativeHistory: &m})
	var decoded agentbridge.NativeHistoryMessage
	if !ok || req.EventType != assignment.EventRiidoLog || req.Metadata["lease_token"] != "lease" || json.Unmarshal([]byte(req.Metadata[agentbridge.NativeHistoryMetadataKey]), &decoded) != nil || decoded != m {
		t.Fatalf("bad native event: %+v", req)
	}
}

func TestNativeHistoryMappingRejectsInvalidTimestamp(t *testing.T) {
	m := agentbridge.NativeHistoryMessage{ObservedAt: time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)}
	if _, ok := eventRequestFromAgentEvent(assignment.Assignment{}, agentbridge.Event{NativeHistory: &m}); ok {
		t.Fatal("invalid native history timestamp was accepted")
	}
}

func TestNativeHistoryReportRejectsMissingAuthorityAndInvalidPayload(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()
	m := agentbridge.NativeHistoryMessage{ObservedAt: time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)}
	event := agentbridge.Event{Kind: agentbridge.EventProgress, NativeHistory: &m}
	if err := plane.ReportEvent(context.Background(), "missing-execution", event); err == nil {
		t.Fatal("missing authority acknowledged native history")
	}
	fake.enqueue(queuedHTTPAssignment("asn-1", "hello"))
	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatal(err)
	}
	if err := plane.ReportEvent(context.Background(), req.ID, event); err == nil {
		t.Fatal("invalid native history encoding acknowledged")
	}
}
