package main

import (
	"fmt"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func (opts taskTransitionOptions) input() (taskdb.TaskTransitionInput, error) {
	if opts.toState == "" {
		return taskdb.TaskTransitionInput{}, fmt.Errorf("--to is required")
	}
	if opts.eventType == "" {
		return taskdb.TaskTransitionInput{}, fmt.Errorf("--event is required")
	}
	to, err := taskdb.ParseTaskState(opts.toState)
	if err != nil {
		return taskdb.TaskTransitionInput{}, err
	}
	return taskdb.TaskTransitionInput{
		TaskID:  opts.taskID,
		ToState: to,
		Event:   ir.EventType(opts.eventType),
		Actor:   opts.actor,
		Source:  opts.source,
		Reason:  opts.reason,
		Guard: taskdb.TaskMutationGuardInput{
			CommandID:   opts.commandID,
			Provider:    opts.provider,
			DecisionLLM: opts.decisionLLM,
			ApprovalID:  opts.approvalID,
		},
	}, nil
}
