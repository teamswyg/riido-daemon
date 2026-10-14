package taskdbplane

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-contracts/task"
)

func TestClaimTaskSkipsHumanGatedTaskWithoutApproval(t *testing.T) {
	path := writeTaskDB(t, humanApprovalGateDB(task.StateQueued))
	plane := newTestPlane(t, path)

	req, err := plane.ClaimTask(context.Background(), "runtime-1")
	if err != nil {
		t.Fatalf("ClaimTask returned error: %v", err)
	}
	if req != nil {
		t.Fatalf("human-gated task without approval should not be claimed: %+v", req)
	}
	db := loadTaskDB(t, path)
	if got := mustFindTask(t, db, "task-human").State; got != task.StateQueued {
		t.Fatalf("state = %s, want Queued", got)
	}
}
