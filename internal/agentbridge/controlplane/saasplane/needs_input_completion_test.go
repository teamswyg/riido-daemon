package saasplane

import (
	"context"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/metadatakeys"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestCompleteTaskReportsNeedsInputWithoutCompletingAssignment(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.enqueue(assignmentcontract.Assignment{
		ID:              "asn-intent",
		TaskID:          "task-copy",
		AgentID:         "agent-copy",
		RuntimeProvider: "claude",
		State:           assignmentcontract.AssignmentQueued,
		LeaseToken:      "lease-intent",
	})
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "agent-copy", RuntimeProvider: "claude"}})
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:claude")
	if err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	result := agentbridge.Result{
		Status: agentbridge.ResultCompleted,
		Output: "@민준용 님 안녕하세요. 어떤 작업부터 진행할까요?",
	}
	if err := plane.CompleteTask(context.Background(), req.ID, result); err != nil {
		t.Fatalf("CompleteTask: %v", err)
	}

	last := fake.events[len(fake.events)-1]
	assertNeedsInputEvent(t, last)
}

func assertNeedsInputEvent(t *testing.T, event assignmentcontract.AgentEventRequest) {
	t.Helper()
	if event.EventType != assignmentcontract.EventAssignmentStateUpdated ||
		event.State != assignmentcontract.AssignmentRunning {
		t.Fatalf("needs-input event shape = %+v", event)
	}
	if got := event.Metadata[metadatakeys.AssignmentResultStatus.String()]; got != "needs_input" {
		t.Fatalf("result status metadata = %q", got)
	}
	if event.Message == "" || event.AssignmentID == "" {
		t.Fatalf("needs-input event missing identity/message: %+v", event)
	}
}
