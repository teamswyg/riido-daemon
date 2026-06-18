package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-contracts/task"
)

func TestTaskValidateRequiresApprovalBeforeExecution(t *testing.T) {
	path := writeValidateTestDB(t, task.StateValidating)
	marker := filepath.Join(t.TempDir(), "should-not-exist")

	err := runValidateCommand(t, path, validateCommandOptions{
		command: "touch " + marker,
		workdir: t.TempDir(),
	})
	if err == nil {
		t.Fatal("expected missing approval id to be rejected")
	}
	if _, statErr := os.Stat(marker); !os.IsNotExist(statErr) {
		t.Fatalf("validation command should not execute without approval, statErr=%v", statErr)
	}
}
