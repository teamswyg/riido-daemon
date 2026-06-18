package taskdbplane

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func TestClaimTaskTransitionsQueuedRowAndBuildsRequest(t *testing.T) {
	path := writeTaskDB(t, queuedClaimRequestDB())
	plane := newTestPlane(t, path)

	req, err := plane.ClaimTask(context.Background(), "runtime-1")
	if err != nil {
		t.Fatalf("ClaimTask returned error: %v", err)
	}
	assertQueuedClaimRequest(t, req, path)
	assertQueuedClaimDB(t, loadTaskDB(t, path))

	second, err := plane.ClaimTask(context.Background(), "runtime-1")
	if err != nil {
		t.Fatalf("second ClaimTask returned error: %v", err)
	}
	if second != nil {
		t.Fatalf("claimed same task twice: %+v", second)
	}
}

func assertQueuedClaimRequest(t *testing.T, req *bridge.TaskRequest, path string) {
	t.Helper()
	if req == nil || req.ID != "task-1" || req.Provider != "codex" || req.Prompt != "implement the patch" {
		t.Fatalf("unexpected request: %+v", req)
	}
	if req.Metadata["workspace_id"] != "project-1" {
		t.Fatalf("workspace metadata missing: %+v", req.Metadata)
	}
	if req.Metadata[metadataTaskDB] != path || req.Metadata[metadataDocument] != "docs/task.md" {
		t.Fatalf("task metadata mismatch: %+v", req.Metadata)
	}
}

func assertQueuedClaimDB(t *testing.T, db taskdb.TaskDB) {
	t.Helper()
	record := mustFindTask(t, db, "task-1")
	if record.State != task.StateClaimed {
		t.Fatalf("state = %s, want Claimed", record.State)
	}
	wantCommand := commandIDPrefix + "task-1:claim:runtime-1"
	if len(db.CommandReceipts) != 1 || db.CommandReceipts[0].CommandID != wantCommand {
		t.Fatalf("claim receipt mismatch: %+v", db.CommandReceipts)
	}
}
