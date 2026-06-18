package taskdb

import (
	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
)

type TaskMutationGuardInput struct {
	CommandID   string
	Provider    string
	DecisionLLM string
	ApprovalID  string
}

type TaskTransitionInput struct {
	TaskID  string
	ToState task.TaskState
	Event   ir.EventType
	Actor   string
	Source  string
	Reason  string
	Guard   TaskMutationGuardInput
}

type TaskEvidenceInput struct {
	TaskID            string
	Command           string
	ExitCode          int
	Result            string
	Actor             string
	Source            string
	Summary           string
	ValidationGate    string
	ProviderRunID     string
	ProviderRunResult string
	Guard             TaskMutationGuardInput
}
