package main

import (
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func runTask(args []string) error {
	if len(args) < 1 {
		printUsage()
		return fmt.Errorf("missing task command")
	}
	taskDBPath, err := taskdb.DefaultTaskDBPath()
	if err != nil {
		return err
	}

	switch taskCommand(args[0]) {
	case taskCommandList:
		return runTaskList(args[1:], taskDBPath)
	case taskCommandTransition:
		return runTaskTransition(args[1:], taskDBPath)
	case taskCommandEvidence:
		return runTaskEvidence(args[1:], taskDBPath)
	case taskCommandValidate:
		return runTaskValidate(args[1:], taskDBPath)
	default:
		printUsage()
		return fmt.Errorf("unknown task command: %s", args[0])
	}
}

func runTaskList(args []string, taskDBPath string) error {
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--task-db":
			value, err := cliRequiredArg(args, &index, "--task-db", "path")
			if err != nil {
				return err
			}
			taskDBPath = value
		case "--help", "-h":
			printUsage()
			return nil
		default:
			return fmt.Errorf("unknown argument: %s", args[index])
		}
	}
	db, err := taskdb.LoadTaskDBOrEmpty(taskDBPath)
	if err != nil {
		return err
	}
	return printJSON(db)
}
