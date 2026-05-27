package taskdb

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
)

func TestGuardedTransitionRequiresApprovalForHumanGatedTask(t *testing.T) {
	db := sampleTaskDB()

	_, _, _, err := ApplyGuardedTaskTransition(db, TaskTransitionInput{
		TaskID:  "task-1",
		ToState: task.StateQueued,
		Event:   ir.EventTaskQueued,
		Actor:   "human",
		Source:  "test",
		Reason:  "ready",
		Guard: TaskMutationGuardInput{
			CommandID: "command:test:queue",
			Provider:  "codex",
		},
	}, fixedTime())
	if err == nil || !strings.Contains(err.Error(), "requires approval_id") {
		t.Fatalf("expected approval guard rejection, got %v", err)
	}
}

func TestGuardedTransitionReplaysCommandIDWithoutDuplicateMutation(t *testing.T) {
	input := TaskTransitionInput{
		TaskID:  "task-1",
		ToState: task.StateQueued,
		Event:   ir.EventTaskQueued,
		Actor:   "human",
		Source:  "test",
		Reason:  "approved",
		Guard: TaskMutationGuardInput{
			CommandID:   "command:test:queue",
			Provider:    "codex",
			DecisionLLM: "codex",
			ApprovalID:  "approval-1",
		},
	}
	first, transition, receipt, err := ApplyGuardedTaskTransition(sampleTaskDB(), input, fixedTime())
	if err != nil {
		t.Fatalf("first transition returned error: %v", err)
	}

	replayed, replayedTransition, replayedReceipt, err := ApplyGuardedTaskTransition(first, input, fixedTime().Add(time.Minute))
	if err != nil {
		t.Fatalf("replayed transition returned error: %v", err)
	}
	if replayedTransition.ID != transition.ID || replayedReceipt.ID != receipt.ID {
		t.Fatalf("replay should return original records: %s/%s vs %s/%s", transition.ID, receipt.ID, replayedTransition.ID, replayedReceipt.ID)
	}
	if len(replayed.Transitions) != len(first.Transitions) || len(replayed.CommandReceipts) != len(first.CommandReceipts) {
		t.Fatalf("replay appended records: first=%d/%d replay=%d/%d", len(first.Transitions), len(first.CommandReceipts), len(replayed.Transitions), len(replayed.CommandReceipts))
	}
}

func TestGuardedEvidencePersistsDeterministicResultAndSaveLoad(t *testing.T) {
	db, _, _, err := ApplyGuardedTaskTransition(sampleTaskDB(), TaskTransitionInput{
		TaskID:  "task-1",
		ToState: task.StateQueued,
		Event:   ir.EventTaskQueued,
		Actor:   "human",
		Source:  "test",
		Reason:  "approved",
		Guard: TaskMutationGuardInput{
			CommandID:   "command:test:queue",
			Provider:    "codex",
			DecisionLLM: "codex",
			ApprovalID:  "approval-1",
		},
	}, fixedTime())
	if err != nil {
		t.Fatalf("queue transition returned error: %v", err)
	}

	updated, evidence, receipt, err := AddGuardedTaskEvidence(db, TaskEvidenceInput{
		TaskID:   "task-1",
		Command:  "go test ./...",
		ExitCode: 0,
		Actor:    "daemon",
		Source:   "test",
		Summary:  "domain gate passed",
		Guard: TaskMutationGuardInput{
			CommandID:   "command:test:evidence",
			Provider:    "codex",
			DecisionLLM: "codex",
			ApprovalID:  "approval-1",
		},
	}, fixedTime().Add(time.Minute))
	if err != nil {
		t.Fatalf("AddGuardedTaskEvidence returned error: %v", err)
	}
	if evidence.Result != "passed" || evidence.ValidationGate != TaskEvidenceValidationV1 || receipt.Result != "passed" {
		t.Fatalf("unexpected evidence/receipt: evidence=%+v receipt=%+v", evidence, receipt)
	}

	path := filepath.Join(t.TempDir(), "task-db.json")
	if err := SaveTaskDB(path, updated); err != nil {
		t.Fatalf("SaveTaskDB returned error: %v", err)
	}
	loaded, err := LoadTaskDB(path)
	if err != nil {
		t.Fatalf("LoadTaskDB returned error: %v", err)
	}
	if loaded.SchemaVersion != TaskDBSchemaVersion || len(loaded.Evidence) != 1 || len(loaded.CommandReceipts) != 2 {
		t.Fatalf("loaded task DB mismatch: %+v", loaded)
	}
}

func sampleTaskDB() TaskDB {
	return TaskDB{
		SchemaVersion:          TaskDBSchemaVersion,
		DecisionGate:           "human-approval-required",
		RecommendedProvider:    "codex",
		RecommendedDecisionLLM: "codex",
		ProviderCandidates: []ProviderCandidate{
			{ID: "codex", Available: true},
		},
		Tasks: []TaskRecord{{
			ID:                     "task-1",
			ProjectID:              "project-1",
			State:                  task.StateCreated,
			Title:                  "Implement task",
			RecommendedProvider:    "codex",
			RecommendedDecisionLLM: "codex",
			RequiresHumanApproval:  true,
			SourceDocumentID:       "doc-1",
			SourceDocumentPath:     "docs/task.md",
		}},
	}
}

func fixedTime() time.Time {
	return time.Date(2026, 5, 28, 1, 2, 3, 0, time.UTC)
}
