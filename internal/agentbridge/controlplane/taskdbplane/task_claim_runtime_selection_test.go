package taskdbplane

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func TestClaimTaskSelectsRuntimePoolCandidate(t *testing.T) {
	path := writeTaskDB(t, singleQueuedCodexTaskDB())
	plane := newTestPlane(t, path)
	registerRuntimeForProvider(t, plane, "runtime-full", "codex", 1, 1)
	registerRuntimeForProvider(t, plane, "runtime-free", "codex", 2, 0)

	first, err := plane.ClaimTask(context.Background(), "runtime-full")
	if err != nil {
		t.Fatalf("ClaimTask runtime-full returned error: %v", err)
	}
	if first != nil {
		t.Fatalf("slot-exhausted runtime should not claim selected task: %+v", first)
	}
	if got := mustFindTask(t, loadTaskDB(t, path), "codex-task").State; got != task.StateQueued {
		t.Fatalf("task should remain queued after non-selected runtime claim, got %s", got)
	}

	second, err := plane.ClaimTask(context.Background(), "runtime-free")
	if err != nil {
		t.Fatalf("ClaimTask runtime-free returned error: %v", err)
	}
	if second == nil || second.ID != "codex-task" {
		t.Fatalf("selected runtime should claim task, got %+v", second)
	}
	if got := mustFindTask(t, loadTaskDB(t, path), "codex-task").State; got != task.StateClaimed {
		t.Fatalf("task should be claimed by selected runtime, got %s", got)
	}
}

func singleQueuedCodexTaskDB() taskdb.TaskDB {
	return taskdb.TaskDB{
		SchemaVersion:       taskdb.TaskDBSchemaVersion,
		RecommendedProvider: "codex",
		ProviderCandidates: []taskdb.ProviderCandidate{
			{ID: "codex", Available: true},
		},
		Tasks: []taskdb.TaskRecord{{
			ID:                  "codex-task",
			ProjectID:           "project-1",
			State:               task.StateQueued,
			Title:               "codex task",
			RecommendedProvider: "codex",
		}},
	}
}
