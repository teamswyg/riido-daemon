package main

import (
	"fmt"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func runTaskTransition(args []string, taskDBPath string) error {
	if len(args) < 1 {
		return fmt.Errorf("task transition requires a task id")
	}
	taskID := args[0]
	toState, eventType := "", ""
	actor, source := "human", "riido-cli"
	reason, provider, decisionLLM, approvalID, commandID := "", "", "", "", ""

	for index := 1; index < len(args); index++ {
		var err error
		switch args[index] {
		case "--task-db":
			taskDBPath, err = cliRequiredArg(args, &index, "--task-db", "path")
		case "--to":
			toState, err = cliRequiredArg(args, &index, "--to", "state")
		case "--event":
			eventType, err = cliRequiredArg(args, &index, "--event", "event type")
		case "--actor":
			actor, err = cliRequiredArg(args, &index, "--actor", "value")
		case "--source":
			source, err = cliRequiredArg(args, &index, "--source", "value")
		case "--reason":
			reason, err = cliRequiredArg(args, &index, "--reason", "value")
		case "--provider":
			provider, err = cliRequiredArg(args, &index, "--provider", "value")
		case "--decision-llm":
			decisionLLM, err = cliRequiredArg(args, &index, "--decision-llm", "value")
		case "--approval-id":
			approvalID, err = cliRequiredArg(args, &index, "--approval-id", "value")
		case "--command-id":
			commandID, err = cliRequiredArg(args, &index, "--command-id", "value")
		case "--help", "-h":
			printUsage()
			return nil
		default:
			return fmt.Errorf("unknown argument: %s", args[index])
		}
		if err != nil {
			return err
		}
	}
	if toState == "" {
		return fmt.Errorf("--to is required")
	}
	if eventType == "" {
		return fmt.Errorf("--event is required")
	}
	to, err := taskdb.ParseTaskState(toState)
	if err != nil {
		return err
	}
	db, err := taskdb.LoadTaskDB(taskDBPath)
	if err != nil {
		return err
	}
	updated, transition, receipt, err := taskdb.ApplyGuardedTaskTransition(db, taskdb.TaskTransitionInput{
		TaskID:  taskID,
		ToState: to,
		Event:   ir.EventType(eventType),
		Actor:   actor,
		Source:  source,
		Reason:  reason,
		Guard: taskdb.TaskMutationGuardInput{
			CommandID:   commandID,
			Provider:    provider,
			DecisionLLM: decisionLLM,
			ApprovalID:  approvalID,
		},
	}, time.Now())
	if err != nil {
		return err
	}
	if err := taskdb.SaveTaskDB(taskDBPath, updated); err != nil {
		return err
	}
	return printJSON(struct {
		OK         bool                            `json:"ok"`
		TaskDBPath string                          `json:"task_db_path"`
		Transition taskdb.TaskTransitionRecord     `json:"transition"`
		Receipt    taskdb.TaskCommandReceiptRecord `json:"receipt"`
	}{
		OK:         true,
		TaskDBPath: taskDBPath,
		Transition: transition,
		Receipt:    receipt,
	})
}
