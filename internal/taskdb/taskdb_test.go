package taskdb

import (
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
