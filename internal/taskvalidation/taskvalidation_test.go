package taskvalidation

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
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
	if result.Validation.Result != "passed" || result.Evidence.Result != "passed" {
		t.Fatalf("validation should pass: %#v", result)
	}
	if result.Task.State != task.StatePatchReady {
		t.Fatalf("validating task should move to patch-ready: %s", result.Task.State)
	}
	if result.Transition == nil || result.Transition.EventType != ir.EventValidationPassed {
		t.Fatalf("missing validation transition: %#v", result.Transition)
	}
	if result.Receipt.CommandID != "command:test:validate" {
		t.Fatalf("evidence receipt command id mismatch: %#v", result.Receipt)
	}
	if result.TransitionReceipt == nil || result.TransitionReceipt.CommandID != "command:test:validate:transition" {
		t.Fatalf("transition receipt command id mismatch: %#v", result.TransitionReceipt)
	}
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

func validationTaskDB(state task.TaskState) taskdb.TaskDB {
	db := taskdb.EmptyTaskDB()
	db.UpdatedAt = "2026-05-20T08:00:00Z"
	db.RecommendedProvider = "codex"
	db.RecommendedDecisionLLM = "codex"
	db.DecisionGate = "human-approval-required"
	db.ProviderCandidates = []taskdb.ProviderCandidate{
		{ID: "codex", SourceWorkflow: "provider-selection", Available: true, ApprovalRequired: true},
	}
	db.Tasks = []taskdb.TaskRecord{
		{
			ID:                     "task:test",
			ProjectID:              "macmini-workspace",
			State:                  state,
			SourceDocumentID:       "mws.test",
			SourceDocumentPath:     "docs/TEST.md",
			Title:                  "test",
			RecommendedProvider:    "codex",
			RecommendedDecisionLLM: "codex",
			RequiresHumanApproval:  true,
			CreatedAt:              "2026-05-20T08:00:00Z",
			UpdatedAt:              "2026-05-20T08:00:00Z",
		},
	}
	db.Transitions = []taskdb.TaskTransitionRecord{
		{
			ID:         "transition:task:test:created",
			TaskID:     "task:test",
			ToState:    task.StateCreated,
			EventType:  ir.EventTaskCreated,
			Actor:      "riido",
			Source:     "test",
			RecordedAt: "2026-05-20T08:00:00Z",
		},
	}
	return db
}

func fixedTime() time.Time {
	return time.Date(2026, 5, 20, 8, 0, 0, 0, time.UTC)
}
