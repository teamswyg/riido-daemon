package main

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func writeValidateTestDB(t *testing.T, state task.TaskState) string {
	t.Helper()

	now := time.Date(2026, 5, 21, 9, 0, 0, 0, time.UTC)
	db := newValidateTestDB(t, state, now)
	path := filepath.Join(t.TempDir(), "task-db.json")
	if err := taskdb.SaveTaskDB(path, db); err != nil {
		t.Fatalf("SaveTaskDB returned error: %v", err)
	}
	return path
}

func newValidateTestDB(
	t *testing.T,
	state task.TaskState,
	now time.Time,
) taskdb.TaskDB {
	t.Helper()

	db := taskdb.EmptyTaskDB()
	db.ProjectionVersion = "test-projection.v1"
	db.Root = t.TempDir()
	db.Domain = "macmini-workspace"
	db.UpdatedAt = now.Format(time.RFC3339Nano)
	db.RecommendedProvider = "codex"
	db.RecommendedDecisionLLM = "codex"
	db.DecisionGate = "human-approval-required"
	db.ProviderCandidates = validateTestProviderCandidates()
	db.Tasks = []taskdb.TaskRecord{newValidateTestTask(state, now)}
	return db
}
