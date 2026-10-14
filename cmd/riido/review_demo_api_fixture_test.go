package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/riidoapi"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func serveReviewDemoCLIAPI(t *testing.T) (string, func()) {
	t.Helper()

	dir := newReviewDemoCLIDir(t)
	socketPath := filepath.Join(dir, "riido.sock")
	taskDBPath := writeReviewDemoCLITaskDB(t, dir)
	ctx, cancel := context.WithCancel(context.Background())
	errs := make(chan error, 1)

	go func() {
		errs <- riidoapi.NewServer(riidoapi.Config{
			AppVersion: "riido-daemon test.v1",
			SocketPath: socketPath,
			TaskDBPath: taskDBPath,
		}).Serve(ctx)
	}()

	waitForReviewDemoCLIAPI(t, socketPath)
	return socketPath, func() {
		stopReviewDemoCLIAPI(t, cancel, errs)
	}
}

func newReviewDemoCLIDir(t *testing.T) string {
	t.Helper()

	dir, err := os.MkdirTemp("", "rcli-")
	if err != nil {
		t.Fatalf("MkdirTemp returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})
	return dir
}

func writeReviewDemoCLITaskDB(t *testing.T, dir string) string {
	t.Helper()

	taskDBPath := filepath.Join(dir, "task-db.json")
	db := taskdb.EmptyTaskDB()
	db.UpdatedAt = "2026-05-26T00:00:00Z"
	if err := taskdb.SaveTaskDB(taskDBPath, db); err != nil {
		t.Fatalf("SaveTaskDB returned error: %v", err)
	}
	return taskDBPath
}
