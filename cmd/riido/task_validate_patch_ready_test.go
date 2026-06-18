package main

import (
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
)

func TestTaskValidateTransitionsValidatingTaskToPatchReady(t *testing.T) {
	path := writeValidateTestDB(t, task.StateValidating)

	err := runValidateCommand(t, path, validateCommandOptions{
		command:    "printf ok",
		workdir:    t.TempDir(),
		approvalID: "approval:test:validate",
	})
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	db := loadValidateCommandDB(t, path)
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
	receipt := findValidateTestReceipt(t, db, "command:test:validate:transition")
	if receipt.Kind != "transition" || receipt.TransitionID != transition.ID {
		t.Fatalf("unexpected transition receipt: %#v", receipt)
	}
}
