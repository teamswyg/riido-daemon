package taskdb

import (
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
)

func queueSampleTask(t *testing.T) TaskDB {
	t.Helper()
	db, _, _, err := ApplyGuardedTaskTransition(sampleTaskDB(), sampleQueueInput(), fixedTime())
	if err != nil {
		t.Fatalf("queue transition returned error: %v", err)
	}
	return db
}

func sampleQueueInput() TaskTransitionInput {
	return TaskTransitionInput{
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
}

func sampleEvidenceInput() TaskEvidenceInput {
	return TaskEvidenceInput{
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
	}
}
