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
	return printTaskEvidenceResult(taskDBPath, evidence, receipt)
}
