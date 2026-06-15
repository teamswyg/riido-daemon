package main

import (
	"context"
	"fmt"
	"time"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
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
	providerForRun, err := validationProviderForTask(db, options.taskID, options.provider)
	if err != nil {
		return err
	}
	if err := validateDecisionLLMForTask(db, options.taskID, options.decisionLLM); err != nil {
		return err
	}
	taskBeforeValidation, ok := findTaskRecord(db, options.taskID)
	if !ok {
		return fmt.Errorf("task %s not found", options.taskID)
	}
	now := time.Now()
	if options.commandID == "" {
		options.commandID = validationCommandID(options.taskID, now)
	}
	result, err := validation.RunCommand(context.Background(), validation.CommandRequest{
		Command:        options.command,
		Workdir:        options.workdir,
		Timeout:        options.timeout,
		CommandID:      options.commandID,
		Provider:       providerForRun,
		ValidationGate: options.validationGate,
		Summary:        options.summary,
	}, now)
	if err != nil {
		return err
	}
	return saveTaskValidationResult(db, taskBeforeValidation, options, providerForRun, result, now)
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

func saveTaskValidationResult(
	db taskdb.TaskDB,
	taskBeforeValidation taskdb.TaskRecord,
	options taskValidateCLI,
	providerForRun string,
	result validation.CommandResult,
	now time.Time,
) error {
	updated, evidence, receipt, err := taskdb.AddGuardedTaskEvidence(db, taskdb.TaskEvidenceInput{
		TaskID:            options.taskID,
		Command:           result.Command,
		ExitCode:          result.ExitCode,
		Result:            result.Result,
		Actor:             options.actor,
		Source:            options.source,
		Summary:           result.Summary,
		ValidationGate:    result.ValidationGate,
		ProviderRunID:     result.ProviderRunID,
		ProviderRunResult: result.ProviderRunResult,
		Guard: taskdb.TaskMutationGuardInput{
			CommandID:   options.commandID,
			Provider:    providerForRun,
			DecisionLLM: options.decisionLLM,
			ApprovalID:  options.approvalID,
		},
	}, now)
	if err != nil {
		return err
	}
	transition, transitionReceipt, err := maybeApplyValidationTransition(
		&updated, taskBeforeValidation, options, providerForRun, result, now,
	)
	if err != nil {
		return err
	}
	if err := taskdb.SaveTaskDB(options.taskDBPath, updated); err != nil {
		return err
	}
	return printJSON(struct {
		OK                bool                             `json:"ok"`
		TaskDBPath        string                           `json:"task_db_path"`
		Validation        validation.CommandResult         `json:"validation"`
		Evidence          taskdb.TaskEvidenceRecord        `json:"evidence"`
		Receipt           taskdb.TaskCommandReceiptRecord  `json:"receipt"`
		Transition        *taskdb.TaskTransitionRecord     `json:"transition,omitempty"`
		TransitionReceipt *taskdb.TaskCommandReceiptRecord `json:"transition_receipt,omitempty"`
	}{
		OK:                evidence.Result == "passed",
		TaskDBPath:        options.taskDBPath,
		Validation:        result,
		Evidence:          evidence,
		Receipt:           receipt,
		Transition:        transition,
		TransitionReceipt: transitionReceipt,
	})
}
