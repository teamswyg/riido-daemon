package main

import (
	"testing"

	"github.com/teamswyg/riido-contracts/task"
)

func TestTaskValidateDoesNotTransitionNonValidatingTask(t *testing.T) {
	path := writeValidateTestDB(t, task.StateCreated)

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
