package main

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func TestTaskListPrintsEmptyDBWhenFileDoesNotExist(t *testing.T) {
	taskDBPath := filepath.Join(t.TempDir(), "missing-task-db.json")

	out := captureStdout(t, func() {
		if err := run([]string{"task", "list", "--task-db", taskDBPath}); err != nil {
			t.Fatalf("run returned error: %v", err)
		}
	})

	var db taskdb.TaskDB
	if err := json.Unmarshal([]byte(out), &db); err != nil {
		t.Fatalf("parse JSON %q: %v", out, err)
	}
	if db.SchemaVersion != taskdb.TaskDBSchemaVersion {
		t.Fatalf("schema_version = %q, want %q", db.SchemaVersion, taskdb.TaskDBSchemaVersion)
	}
	if len(db.Tasks) != 0 {
		t.Fatalf("missing DB should list zero tasks, got %d", len(db.Tasks))
	}
}
