package saasplane

import (
	"context"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestCompleteTaskFallsBackToAccumulatedBody(t *testing.T) {
	fake, plane, req := newTextDeltaFixture(t)
	reportTextDeltas(t, plane, req.ID, "The answer ", "is fully streamed here.")

	if err := plane.CompleteTask(context.Background(), req.ID, agentbridge.Result{
		Status: agentbridge.ResultCompleted,
	}); err != nil {
		t.Fatalf("CompleteTask: %v", err)
	}

	last := fake.events[len(fake.events)-1]
	if last.State != assignmentcontract.AssignmentCompleted {
		t.Fatalf("expected completed state, got %q", last.State)
	}
	if last.Message != "The answer is fully streamed here." {
		t.Fatalf("completion message should fall back to accumulated body, got %q", last.Message)
	}
}
