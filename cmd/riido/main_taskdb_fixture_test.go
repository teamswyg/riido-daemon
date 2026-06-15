package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/riidoapi"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func writeValidateTestDB(t *testing.T, state task.TaskState) string {
	t.Helper()
	now := time.Date(2026, 5, 21, 9, 0, 0, 0, time.UTC)
	db := taskdb.EmptyTaskDB()
	db.ProjectionVersion = "test-projection.v1"
	db.Root = t.TempDir()
	db.Domain = "macmini-workspace"
	db.UpdatedAt = now.Format(time.RFC3339Nano)
	db.RecommendedProvider = "codex"
	db.RecommendedDecisionLLM = "codex"
	db.DecisionGate = "human-approval-required"
	db.ProviderCandidates = []taskdb.ProviderCandidate{
		{ID: "codex", SourceWorkflow: "provider-selection", Available: true, ApprovalRequired: true},
	}
	db.Tasks = []taskdb.TaskRecord{
		{
			ID:                     validateTestTaskID,
			ProjectID:              "macmini-workspace",
			State:                  state,
			SourceDocumentID:       "mws.test",
			SourceDocumentPath:     "docs/TEST.md",
			Title:                  "Validation FSM test",
			Owner:                  "human",
			SourceStatus:           "in-progress",
			RecommendedProvider:    "codex",
			RecommendedDecisionLLM: "codex",
			RequiresHumanApproval:  true,
			HarnessNextDirection:   "top-down",
			CreatedAt:              now.Format(time.RFC3339Nano),
			UpdatedAt:              now.Format(time.RFC3339Nano),
		},
	}
	path := filepath.Join(t.TempDir(), "task-db.json")
	if err := taskdb.SaveTaskDB(path, db); err != nil {
		t.Fatalf("SaveTaskDB returned error: %v", err)
	}
	return path
}

func serveReviewDemoCLIAPI(t *testing.T) (string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "rcli-")
	if err != nil {
		t.Fatalf("MkdirTemp returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})
	socketPath := filepath.Join(dir, "riido.sock")
	taskDBPath := filepath.Join(dir, "task-db.json")
	db := taskdb.EmptyTaskDB()
	db.UpdatedAt = "2026-05-26T00:00:00Z"
	if err := taskdb.SaveTaskDB(taskDBPath, db); err != nil {
		t.Fatalf("SaveTaskDB returned error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	errs := make(chan error, 1)
	go func() {
		errs <- riidoapi.NewServer(riidoapi.Config{SocketPath: socketPath, TaskDBPath: taskDBPath}).Serve(ctx)
	}()
	waitForReviewDemoCLIAPI(t, socketPath)
	return socketPath, func() {
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

func waitForReviewDemoCLIAPI(t *testing.T, socketPath string) {
	t.Helper()
	client := riidoapi.NewClient(socketPath)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		var status riidoapi.Status
		if err := client.Request(context.Background(), "status", nil, &status); err == nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("riido API did not become ready at %s", socketPath)
}

func findValidateTestTask(t *testing.T, db taskdb.TaskDB) taskdb.TaskRecord {
	t.Helper()
	for _, record := range db.Tasks {
		if record.ID == validateTestTaskID {
			return record
		}
	}
	t.Fatalf("task %s not found", validateTestTaskID)
	return taskdb.TaskRecord{}
}

func findValidateTestReceipt(t *testing.T, db taskdb.TaskDB, commandID string) taskdb.TaskCommandReceiptRecord {
	t.Helper()
	for _, receipt := range db.CommandReceipts {
		if receipt.CommandID == commandID {
			return receipt
		}
	}
	t.Fatalf("receipt for command %s not found", commandID)
	return taskdb.TaskCommandReceiptRecord{}
}
