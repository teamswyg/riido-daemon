package main

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

const validateTestTaskID = "task:test"

type validateCommandOptions struct {
	command    string
	workdir    string
	approvalID string
}

func runValidateCommand(t *testing.T, path string, opts validateCommandOptions) error {
	t.Helper()
	args := []string{
		"task", "validate", validateTestTaskID,
		"--task-db", path,
		"--command", opts.command,
		"--workdir", opts.workdir,
		"--command-id", "command:test:validate",
		"--provider", "codex",
		"--decision-llm", "codex",
		"--timeout-seconds", "5",
	}
	if opts.approvalID != "" {
		args = append(args, "--approval-id", opts.approvalID)
	}
	return run(args)
}

func loadValidateCommandDB(t *testing.T, path string) taskdb.TaskDB {
	t.Helper()
	db, err := taskdb.LoadTaskDB(path)
	if err != nil {
		t.Fatalf("LoadTaskDB returned error: %v", err)
	}
	return db
}
