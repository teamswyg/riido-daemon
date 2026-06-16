package main

import (
	"context"
	"fmt"
	"time"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
	"github.com/teamswyg/riido-daemon/internal/taskvalidation"
	"github.com/teamswyg/riido-daemon/internal/validation"
)

type taskValidateCLI struct {
	taskDBPath     string
	taskID         string
	command        string
	workdir        string
	timeout        time.Duration
	actor          string
	source         string
	summary        string
	provider       string
	decisionLLM    string
	approvalID     string
	commandID      string
	validationGate string
}

func runTaskValidate(args []string, taskDBPath string) error {
	options, err := parseTaskValidateCLI(args, taskDBPath)
	if err != nil {
		if isCLIHelp(err) {
			return nil
		}
		return err
	}
	db, err := taskdb.LoadTaskDB(options.taskDBPath)
	if err != nil {
		return err
	}
	now := time.Now()
	result, err := taskvalidation.Run(context.Background(), db, taskvalidation.Request{
		TaskID:         options.taskID,
		Command:        options.command,
		Workdir:        options.workdir,
		Timeout:        options.timeout,
		Actor:          options.actor,
		Source:         options.source,
		Summary:        options.summary,
		Provider:       options.provider,
		DecisionLLM:    options.decisionLLM,
		ApprovalID:     options.approvalID,
		CommandID:      options.commandID,
		ValidationGate: options.validationGate,
	}, now)
	if err != nil {
		return err
	}
	if err := taskdb.SaveTaskDB(options.taskDBPath, result.TaskDB); err != nil {
		return err
	}
	return printTaskValidationResult(options.taskDBPath, result)
}

func parseTaskValidateCLI(args []string, taskDBPath string) (taskValidateCLI, error) {
	if len(args) < 1 {
		return taskValidateCLI{}, fmt.Errorf("task validate requires a task id")
	}
	options := taskValidateCLI{taskDBPath: taskDBPath, taskID: args[0], actor: "daemon", source: "riido-validation-runner"}
	for index := 1; index < len(args); index++ {
		if err := parseTaskValidateFlag(args, &index, &options); err != nil {
			return taskValidateCLI{}, err
		}
	}
	if options.command == "" {
		return taskValidateCLI{}, fmt.Errorf("--command is required")
	}
	if options.approvalID == "" {
		return taskValidateCLI{}, fmt.Errorf("--approval-id is required before validation command execution")
	}
	return options, nil
}

func parseTaskValidateFlag(args []string, index *int, options *taskValidateCLI) error {
	var err error
	switch args[*index] {
	case "--task-db":
		options.taskDBPath, err = cliRequiredArg(args, index, "--task-db", "path")
	case "--command":
		options.command, err = cliRequiredArg(args, index, "--command", "value")
	case "--workdir":
		options.workdir, err = cliRequiredArg(args, index, "--workdir", "path")
	case "--timeout-seconds":
		var seconds int
		seconds, err = cliRequiredInt(args, index, "--timeout-seconds")
		if err == nil && seconds <= 0 {
			err = fmt.Errorf("--timeout-seconds must be positive")
		}
		options.timeout = time.Duration(seconds) * time.Second
	case "--actor":
		options.actor, err = cliRequiredArg(args, index, "--actor", "value")
	case "--source":
		options.source, err = cliRequiredArg(args, index, "--source", "value")
	case "--summary":
		options.summary, err = cliRequiredArg(args, index, "--summary", "value")
	case "--provider":
		options.provider, err = cliRequiredArg(args, index, "--provider", "value")
	case "--decision-llm":
		options.decisionLLM, err = cliRequiredArg(args, index, "--decision-llm", "value")
	case "--approval-id":
		options.approvalID, err = cliRequiredArg(args, index, "--approval-id", "value")
	case "--command-id":
		options.commandID, err = cliRequiredArg(args, index, "--command-id", "value")
	case "--validation-gate":
		options.validationGate, err = cliRequiredArg(args, index, "--validation-gate", "value")
	case "--help", "-h":
		printUsage()
		return errCLIHelp
	default:
		return fmt.Errorf("unknown argument: %s", args[*index])
	}
	return err
}

func printTaskValidationResult(taskDBPath string, result taskvalidation.Result) error {
	return printJSON(struct {
		OK                bool                             `json:"ok"`
		TaskDBPath        string                           `json:"task_db_path"`
		Validation        validation.CommandResult         `json:"validation"`
		Evidence          taskdb.TaskEvidenceRecord        `json:"evidence"`
		Receipt           taskdb.TaskCommandReceiptRecord  `json:"receipt"`
		Transition        *taskdb.TaskTransitionRecord     `json:"transition,omitempty"`
		TransitionReceipt *taskdb.TaskCommandReceiptRecord `json:"transition_receipt,omitempty"`
	}{
		OK:                result.Evidence.Result == "passed",
		TaskDBPath:        taskDBPath,
		Validation:        result.Validation,
		Evidence:          result.Evidence,
		Receipt:           result.Receipt,
		Transition:        result.Transition,
		TransitionReceipt: result.TransitionReceipt,
	})
}
