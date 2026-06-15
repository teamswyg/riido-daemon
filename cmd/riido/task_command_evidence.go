package main

import (
	"fmt"
	"time"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func runTaskEvidence(args []string, taskDBPath string) error {
	if len(args) < 1 {
		return fmt.Errorf("task evidence requires a task id")
	}
	input := taskdb.TaskEvidenceInput{
		TaskID: args[0],
		Actor:  "daemon",
		Source: "riido-cli",
	}

	for index := 1; index < len(args); index++ {
		if err := parseTaskEvidenceFlag(args, &index, &taskDBPath, &input); err != nil {
			if isCLIHelp(err) {
				return nil
			}
			return err
		}
	}
	if input.Command == "" {
		return fmt.Errorf("--command is required")
	}
	db, err := taskdb.LoadTaskDB(taskDBPath)
	if err != nil {
		return err
	}
	updated, evidence, receipt, err := taskdb.AddGuardedTaskEvidence(db, input, time.Now())
	if err != nil {
		return err
	}
	if err := taskdb.SaveTaskDB(taskDBPath, updated); err != nil {
		return err
	}
	return printJSON(struct {
		OK         bool                            `json:"ok"`
		TaskDBPath string                          `json:"task_db_path"`
		Evidence   taskdb.TaskEvidenceRecord       `json:"evidence"`
		Receipt    taskdb.TaskCommandReceiptRecord `json:"receipt"`
	}{
		OK:         true,
		TaskDBPath: taskDBPath,
		Evidence:   evidence,
		Receipt:    receipt,
	})
}

func parseTaskEvidenceFlag(args []string, index *int, taskDBPath *string, input *taskdb.TaskEvidenceInput) error {
	var err error
	switch args[*index] {
	case "--task-db":
		*taskDBPath, err = cliRequiredArg(args, index, "--task-db", "path")
	case "--command":
		input.Command, err = cliRequiredArg(args, index, "--command", "value")
	case "--exit-code":
		input.ExitCode, err = cliRequiredInt(args, index, "--exit-code")
	case "--result":
		input.Result, err = cliRequiredArg(args, index, "--result", "value")
	case "--actor":
		input.Actor, err = cliRequiredArg(args, index, "--actor", "value")
	case "--source":
		input.Source, err = cliRequiredArg(args, index, "--source", "value")
	case "--summary":
		input.Summary, err = cliRequiredArg(args, index, "--summary", "value")
	case "--provider":
		input.Guard.Provider, err = cliRequiredArg(args, index, "--provider", "value")
	case "--decision-llm":
		input.Guard.DecisionLLM, err = cliRequiredArg(args, index, "--decision-llm", "value")
	case "--approval-id":
		input.Guard.ApprovalID, err = cliRequiredArg(args, index, "--approval-id", "value")
	case "--command-id":
		input.Guard.CommandID, err = cliRequiredArg(args, index, "--command-id", "value")
	case "--validation-gate":
		input.ValidationGate, err = cliRequiredArg(args, index, "--validation-gate", "value")
	case "--provider-run-id":
		input.ProviderRunID, err = cliRequiredArg(args, index, "--provider-run-id", "value")
	case "--provider-run-result":
		input.ProviderRunResult, err = cliRequiredArg(args, index, "--provider-run-result", "value")
	case "--help", "-h":
		printUsage()
		return errCLIHelp
	default:
		return fmt.Errorf("unknown argument: %s", args[*index])
	}
	return err
}
