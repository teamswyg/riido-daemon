package workdir

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrepareCreatesTreeAndEnforcesWorkspaceID(t *testing.T) {
	root := t.TempDir()
	a := NewFSAdapter(root)
	_, err := a.Prepare(TaskID{Task: "task-1"})
	if err == nil {
		t.Fatal("expected error for empty workspace id")
	}
	ws, err := a.Prepare(TaskID{Workspace: "ws-1", Task: "task-1", Run: "run-1"})
	if err != nil {
		t.Fatal(err)
	}
	wantRoot := filepath.Join(root, "ws-1", "tasks", "task-1", "runs", "run-1")
	if ws.Root != wantRoot {
		t.Fatalf("workspace root = %q, want %q", ws.Root, wantRoot)
	}
	assertWorkspaceDirs(t, ws)
	assertGCMeta(t, ws)
}

func assertWorkspaceDirs(t *testing.T, ws Workspace) {
	t.Helper()
	for _, sub := range []string{"workdir", "output", "logs", "artifacts", "native-config", "ir"} {
		info, err := os.Stat(filepath.Join(ws.Root, sub))
		if err != nil || !info.IsDir() {
			t.Fatalf("expected %s subdir: info=%+v err=%v", sub, info, err)
		}
	}
}

func assertGCMeta(t *testing.T, ws Workspace) {
	t.Helper()
	meta, err := os.ReadFile(filepath.Join(ws.Root, ".gc_meta.json"))
	if err != nil {
		t.Fatalf("missing gc meta: %v", err)
	}
	for _, want := range []string{`"workspace_id":"ws-1"`, `"task_id":"task-1"`, `"run_id":"run-1"`} {
		if !strings.Contains(string(meta), want) {
			t.Fatalf("gc meta missing %q:\n%s", want, meta)
		}
	}
}
