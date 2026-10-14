package workdir

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestCleanupArchivedBeforeRemovesOnlyExpiredArchivedRuns(t *testing.T) {
	a := NewFSAdapter(t.TempDir())
	oldRun := preparedRun(t, a, "task-old", "run-old")
	freshRun := preparedRun(t, a, "task-fresh", "run-fresh")
	activeRun := preparedRun(t, a, "task-active", "run-active")
	cutoff := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	archiveRun(t, a, oldRun, "completed", cutoff.Add(-time.Hour))
	archiveRun(t, a, freshRun, "failed", cutoff.Add(time.Hour))

	result, err := a.CleanupArchivedBefore(context.Background(), CleanupRequest{
		ArchivedBefore: cutoff,
		RemovedAt:      cutoff.Add(2 * time.Hour),
	})
	if err != nil {
		t.Fatalf("CleanupArchivedBefore: %v", err)
	}
	if result.ScannedArchiveRecords != 2 || len(result.Removed) != 1 {
		t.Fatalf("cleanup result = %+v", result)
	}
	if result.Removed[0].RunRoot != oldRun.Root || result.Removed[0].Archive.ResultStatus != "completed" {
		t.Fatalf("removed record = %+v", result.Removed[0])
	}
	assertRunRemoved(t, oldRun)
	assertRunKept(t, freshRun)
	assertRunKept(t, activeRun)
}

func TestCleanupArchivedBeforeRequiresCutoff(t *testing.T) {
	_, err := NewFSAdapter(t.TempDir()).CleanupArchivedBefore(context.Background(), CleanupRequest{})
	if err == nil {
		t.Fatal("expected error for empty cleanup cutoff")
	}
}

func assertRunRemoved(t *testing.T, ws Workspace) {
	t.Helper()
	if _, err := os.Stat(ws.Root); !os.IsNotExist(err) {
		t.Fatalf("run should be removed %s: stat err=%v", ws.Root, err)
	}
}
