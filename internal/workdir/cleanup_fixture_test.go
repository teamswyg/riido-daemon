package workdir

import (
	"os"
	"testing"
	"time"
)

func preparedRun(t *testing.T, a *FSAdapter, task, run string) Workspace {
	t.Helper()
	ws, err := a.Prepare(TaskID{Workspace: "ws-1", Task: task, Run: run})
	if err != nil {
		t.Fatal(err)
	}
	return ws
}

func archiveRun(t *testing.T, a *FSAdapter, ws Workspace, status string, archivedAt time.Time) {
	t.Helper()
	if _, err := a.Archive(ws, ArchiveRequest{ResultStatus: status, ArchivedAt: archivedAt}); err != nil {
		t.Fatal(err)
	}
}

func assertRunKept(t *testing.T, ws Workspace) {
	t.Helper()
	if info, err := os.Stat(ws.Root); err != nil || !info.IsDir() {
		t.Fatalf("run should remain %s: info=%+v err=%v", ws.Root, info, err)
	}
}
