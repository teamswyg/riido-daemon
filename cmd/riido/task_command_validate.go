package main

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
	"github.com/teamswyg/riido-daemon/internal/taskvalidation"
)

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
