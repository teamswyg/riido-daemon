package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/riidoapi"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

const validateTestTaskID = "task:test"

func TestTopLevelHelpReturnsNil(t *testing.T) {
	if err := run([]string{"--help"}); err != nil {
		t.Fatalf("help returned error: %v", err)
	}
}

func TestTaskValidateTransitionsValidatingTaskToPatchReady(t *testing.T) {
	taskDBPath := writeValidateTestDB(t, task.StateValidating)

	err := run([]string{
		"task", "validate", validateTestTaskID,
		"--task-db", taskDBPath,
		"--command", "printf ok",
		"--workdir", t.TempDir(),
		"--approval-id", "approval:test:validate",
		"--command-id", "command:test:validate",
		"--provider", "codex",
		"--decision-llm", "codex",
		"--timeout-seconds", "5",
	})
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	db, err := taskdb.LoadTaskDB(taskDBPath)
	if err != nil {
		t.Fatalf("LoadTaskDB returned error: %v", err)
	}
	record := findValidateTestTask(t, db)
	if record.State != task.StatePatchReady {
		t.Fatalf("expected PatchReady state, got %s", record.State)
	}
	if len(db.Evidence) != 1 || db.Evidence[0].Result != "passed" {
		t.Fatalf("expected one passed evidence record, got %#v", db.Evidence)
	}
	if len(db.Transitions) != 1 {
		t.Fatalf("expected one validation transition, got %d", len(db.Transitions))
	}
	transition := db.Transitions[0]
	if transition.FromState != task.StateValidating || transition.ToState != task.StatePatchReady || transition.EventType != ir.EventValidationPassed {
		t.Fatalf("unexpected validation transition: %#v", transition)
	}
	if len(db.CommandReceipts) != 2 {
		t.Fatalf("expected evidence and transition receipts, got %d", len(db.CommandReceipts))
	}
	if db.CommandReceipts[0].CommandID == db.CommandReceipts[1].CommandID {
		t.Fatalf("evidence and transition receipts should have distinct command IDs: %#v", db.CommandReceipts)
	}
	transitionReceipt := findValidateTestReceipt(t, db, "command:test:validate:transition")
	if transitionReceipt.Kind != "transition" || transitionReceipt.TransitionID != transition.ID {
		t.Fatalf("unexpected transition receipt: %#v", transitionReceipt)
	}
}

func TestTaskValidateTransitionsValidatingTaskToFailed(t *testing.T) {
	taskDBPath := writeValidateTestDB(t, task.StateValidating)

	err := run([]string{
		"task", "validate", validateTestTaskID,
		"--task-db", taskDBPath,
		"--command", "exit 7",
		"--workdir", t.TempDir(),
		"--approval-id", "approval:test:validate",
		"--command-id", "command:test:validate",
		"--provider", "codex",
		"--decision-llm", "codex",
		"--timeout-seconds", "5",
	})
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	db, err := taskdb.LoadTaskDB(taskDBPath)
	if err != nil {
		t.Fatalf("LoadTaskDB returned error: %v", err)
	}
	record := findValidateTestTask(t, db)
	if record.State != task.StateFailed {
		t.Fatalf("expected Failed state, got %s", record.State)
	}
	if len(db.Evidence) != 1 || db.Evidence[0].Result != "failed" || db.Evidence[0].ExitCode != 7 {
		t.Fatalf("expected one failed evidence record, got %#v", db.Evidence)
	}
	if len(db.Transitions) != 1 {
		t.Fatalf("expected one validation transition, got %d", len(db.Transitions))
	}
	transition := db.Transitions[0]
	if transition.FromState != task.StateValidating || transition.ToState != task.StateFailed || transition.EventType != ir.EventValidationFailed {
		t.Fatalf("unexpected validation transition: %#v", transition)
	}
}

func TestTaskValidateDoesNotTransitionNonValidatingTask(t *testing.T) {
	taskDBPath := writeValidateTestDB(t, task.StateCreated)

	err := run([]string{
		"task", "validate", validateTestTaskID,
		"--task-db", taskDBPath,
		"--command", "printf ok",
		"--workdir", t.TempDir(),
		"--approval-id", "approval:test:validate",
		"--command-id", "command:test:validate",
		"--provider", "codex",
		"--decision-llm", "codex",
		"--timeout-seconds", "5",
	})
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	db, err := taskdb.LoadTaskDB(taskDBPath)
	if err != nil {
		t.Fatalf("LoadTaskDB returned error: %v", err)
	}
	record := findValidateTestTask(t, db)
	if record.State != task.StateCreated {
		t.Fatalf("expected Created state to be preserved, got %s", record.State)
	}
	if len(db.Evidence) != 1 {
		t.Fatalf("expected one evidence record, got %d", len(db.Evidence))
	}
	if len(db.Transitions) != 0 {
		t.Fatalf("non-Validating validation should not append transitions: %#v", db.Transitions)
	}
	if len(db.CommandReceipts) != 1 || db.CommandReceipts[0].Kind != "evidence" {
		t.Fatalf("expected only an evidence receipt, got %#v", db.CommandReceipts)
	}
}

func TestTaskValidateRequiresApprovalBeforeExecution(t *testing.T) {
	taskDBPath := writeValidateTestDB(t, task.StateValidating)
	marker := filepath.Join(t.TempDir(), "should-not-exist")

	err := run([]string{
		"task", "validate", validateTestTaskID,
		"--task-db", taskDBPath,
		"--command", "touch " + marker,
		"--workdir", t.TempDir(),
		"--command-id", "command:test:validate",
		"--provider", "codex",
		"--decision-llm", "codex",
		"--timeout-seconds", "5",
	})
	if err == nil {
		t.Fatal("expected missing approval id to be rejected")
	}
	if _, statErr := os.Stat(marker); !os.IsNotExist(statErr) {
		t.Fatalf("validation command should not execute without approval, statErr=%v", statErr)
	}
}

func TestAPIReviewDemoCommandUsesLocalControlSurface(t *testing.T) {
	socketPath, stop := serveReviewDemoCLIAPI(t)
	defer stop()

	err := run([]string{
		"api", "review-demo",
		"--socket", socketPath,
		"--channel", "mac-app-store",
		"--review-demo-consent-granted", "true",
	})
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}
}

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
