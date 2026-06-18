package taskvalidation

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-contracts/task"
)

func TestRunRecordsEvidenceAndValidationTransition(t *testing.T) {
	now := fixedTime()
	result, err := Run(context.Background(), validationTaskDB(task.StateValidating), Request{
		TaskID:      "task:test",
		Command:     "printf ok",
		Workdir:     t.TempDir(),
		Provider:    "codex",
		DecisionLLM: "codex",
		ApprovalID:  "approval:test:validate",
		CommandID:   "command:test:validate",
	}, now)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	assertValidationPassedRun(t, result, "command:test:validate")
}

func TestRunRejectsMissingApprovalBeforeCommandExecution(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "should-not-exist")
	_, err := Run(context.Background(), validationTaskDB(task.StateValidating), Request{
		TaskID:  "task:test",
		Command: "touch " + marker,
		Workdir: t.TempDir(),
	}, fixedTime())
	if err == nil {
		t.Fatal("expected missing approval to fail")
	}
	if _, statErr := os.Stat(marker); !os.IsNotExist(statErr) {
		t.Fatalf("validation command should not execute without approval, statErr=%v", statErr)
	}
}
