package taskdbplane

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

const codexTaskID = "codex-task"

func writeQueuedCodexTaskDB(t *testing.T) string {
	t.Helper()
	return writeTaskDB(t, queuedCodexTaskDB())
}

func queuedCodexTaskDB() taskdb.TaskDB {
	return taskdb.TaskDB{
		SchemaVersion:       taskdb.TaskDBSchemaVersion,
		RecommendedProvider: "codex",
		ProviderCandidates: []taskdb.ProviderCandidate{
			{ID: "codex", Available: true},
		},
		Tasks: []taskdb.TaskRecord{{
			ID:                  codexTaskID,
			State:               task.StateQueued,
			Title:               "codex task",
			RecommendedProvider: "codex",
		}},
	}
}

func claimRuntimeTask(t *testing.T, plane *Plane, runtimeID string) *bridge.TaskRequest {
	t.Helper()
	req, err := plane.ClaimTask(context.Background(), runtimeID)
	if err != nil || req == nil {
		t.Fatalf("ClaimTask = %+v, %v", req, err)
	}
	return req
}

func assertTaskState(
	t *testing.T,
	path string,
	id string,
	want task.TaskState,
) {
	t.Helper()
	if got := mustFindTask(t, loadTaskDB(t, path), id).State; got != want {
		t.Fatalf("state should remain %s, got %s", want, got)
	}
}
