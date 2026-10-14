package main

import (
	"fmt"
	"time"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func runTaskTransition(args []string, taskDBPath string) error {
	if len(args) < 1 {
		return fmt.Errorf("task transition requires a task id")
	}
	opts, err := parseTaskTransitionOptions(args, taskDBPath)
	if err != nil {
		if isCLIHelp(err) {
			return nil
		}
		return err
	}
	taskDBPath = opts.taskDBPath

	db, err := taskdb.LoadTaskDB(taskDBPath)
	if err != nil {
		return err
	}
	input, err := opts.input()
	if err != nil {
		return err
	}
	updated, transition, receipt, err := taskdb.ApplyGuardedTaskTransition(db, input, time.Now())
	if err != nil {
		return err
	}
	if err := taskdb.SaveTaskDB(taskDBPath, updated); err != nil {
		return err
	}
	return printTaskTransitionResult(taskDBPath, transition, receipt)
}
