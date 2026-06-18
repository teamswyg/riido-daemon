package riidoapi

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func serveTestAPI(t *testing.T) (string, string, func()) {
	t.Helper()
	return serveTestAPIWithState(t, task.StateCreated)
}

func serveTestAPIWithState(t *testing.T, state task.TaskState) (string, string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "rapi-")
	if err != nil {
		t.Fatalf("MkdirTemp failed: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	socketPath := filepath.Join(dir, "riido.sock")
	taskDBPath := filepath.Join(dir, "task-db.json")
	if err := taskdb.SaveTaskDB(taskDBPath, seedTaskDB(state)); err != nil {
		t.Fatalf("SaveTaskDB failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	errs := make(chan error, 1)
	go func() {
		errs <- NewServer(Config{SocketPath: socketPath, TaskDBPath: taskDBPath}).Serve(ctx)
	}()
	waitForAPI(t, socketPath)
	return socketPath, taskDBPath, stopTestAPI(t, cancel, errs)
}

func stopTestAPI(t *testing.T, cancel context.CancelFunc, errs <-chan error) func() {
	t.Helper()
	return func() {
		cancel()
		select {
		case err := <-errs:
			if err != nil {
				t.Fatalf("Serve returned error: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for API server shutdown")
		}
	}
}
