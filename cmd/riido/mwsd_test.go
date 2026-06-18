package main

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/project"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func TestMwsdSyncWritesStateAndTaskDB(t *testing.T) {
	socketPath, stop := serveTestMwsd(t)
	defer stop()
	statePath := filepath.Join(t.TempDir(), "workspace-state.json")
	taskDBPath := filepath.Join(t.TempDir(), "task-db.json")

	out := captureStdout(t, func() {
		if err := run([]string{
			"mwsd", "sync",
			"--socket", socketPath,
			"--state", statePath,
			"--task-db", taskDBPath,
		}); err != nil {
			t.Fatalf("run mwsd sync: %v", err)
		}
	})
	if !json.Valid([]byte(out)) {
		t.Fatalf("sync output is not JSON: %s", out)
	}
	state, err := project.LoadState(statePath)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	if len(state.Tasks) != 1 || state.Tasks[0].ID != "task:mws.cli" {
		t.Fatalf("unexpected state tasks: %#v", state.Tasks)
	}
	db, err := taskdb.LoadTaskDB(taskDBPath)
	if err != nil {
		t.Fatalf("LoadTaskDB: %v", err)
	}
	if len(db.Tasks) != 1 || db.Tasks[0].ID != "task:mws.cli" {
		t.Fatalf("unexpected task DB tasks: %#v", db.Tasks)
	}
	if len(db.Transitions) != 1 {
		t.Fatalf("expected one created transition, got %d", len(db.Transitions))
	}
}
