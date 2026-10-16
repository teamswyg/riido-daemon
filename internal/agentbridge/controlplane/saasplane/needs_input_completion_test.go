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
	if event.Message == "" || event.AssignmentID != "asn-intent" {
		t.Fatalf("needs-input event missing identity/message: %+v", event)
	}
}

// A whitespace final must preserve a streamed clarification as needs_input,
// rather than accepting an empty success or classifying a real question as failure.
func TestCompleteTaskPreservesStreamedClarification(t *testing.T) {
	fake, plane, req := newTextDeltaFixture(t)
	question := "어떤 작업부터 진행할까요?"
	reportTextDeltas(t, plane, req.ID, question)
	if err := plane.CompleteTask(context.Background(), req.ID, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: " \n"}); err != nil {
		t.Fatal(err)
	}
	last := fake.events[len(fake.events)-1]
	if last.State != assignmentcontract.AssignmentRunning || last.Message != question || last.Metadata[metadatakeys.AssignmentResultStatus.String()] != "needs_input" {
		t.Fatalf("streamed clarification changed semantics: %+v", last)
	}
}
