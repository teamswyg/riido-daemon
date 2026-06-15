package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
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
