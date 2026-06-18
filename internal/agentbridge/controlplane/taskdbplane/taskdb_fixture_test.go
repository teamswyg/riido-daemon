package taskdbplane

import (
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func writeTaskDB(t *testing.T, db taskdb.TaskDB) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "task-db.json")
	if err := taskdb.SaveTaskDB(path, db); err != nil {
		t.Fatalf("SaveTaskDB: %v", err)
	}
	return path
}

func loadTaskDB(t *testing.T, path string) taskdb.TaskDB {
	t.Helper()
	db, err := taskdb.LoadTaskDB(path)
	if err != nil {
		t.Fatalf("LoadTaskDB: %v", err)
	}
	return db
}

func claimedTaskDB() taskdb.TaskDB {
	return taskdb.TaskDB{
		SchemaVersion:       taskdb.TaskDBSchemaVersion,
		RecommendedProvider: "codex",
		ProviderCandidates: []taskdb.ProviderCandidate{
			{ID: "codex", Available: true},
		},
		Tasks: []taskdb.TaskRecord{{
			ID:                  "task-1",
			ProjectID:           "project-1",
			State:               task.StateClaimed,
			Title:               "run it",
			RecommendedProvider: "codex",
		}},
	}
}

func mustFindTask(t *testing.T, db taskdb.TaskDB, id string) taskdb.TaskRecord {
	t.Helper()
	record, ok := findTask(db, id)
	if !ok {
		t.Fatalf("task %s not found", id)
	}
	return record
}

func assertTransition(t *testing.T, db taskdb.TaskDB, event ir.EventType) {
	t.Helper()
	for _, transition := range db.Transitions {
		if transition.EventType == event {
			return
		}
	}
	t.Fatalf("transition %s not found in %+v", event, db.Transitions)
}
