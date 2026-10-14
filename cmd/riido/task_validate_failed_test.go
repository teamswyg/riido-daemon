package main

import (
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
)

func TestTaskValidateTransitionsValidatingTaskToFailed(t *testing.T) {
	path := writeValidateTestDB(t, task.StateValidating)

	err := runValidateCommand(t, path, validateCommandOptions{
		command:    "exit 7",
		workdir:    t.TempDir(),
		approvalID: "approval:test:validate",
	})
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	db := loadValidateCommandDB(t, path)
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
