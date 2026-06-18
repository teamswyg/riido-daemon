package main

import "fmt"

func parseTaskValidateCLI(args []string, taskDBPath string) (taskValidateCLI, error) {
	if len(args) < 1 {
		return taskValidateCLI{}, fmt.Errorf("task validate requires a task id")
	}
	options := defaultTaskValidateCLI(taskDBPath, args[0])
	for index := 1; index < len(args); index++ {
		if err := parseTaskValidateFlag(args, &index, &options); err != nil {
			return taskValidateCLI{}, err
		}
	}
	if err := validateTaskValidateCLI(options); err != nil {
		return taskValidateCLI{}, err
	}
	return options, nil
}

func validateTaskValidateCLI(options taskValidateCLI) error {
	if options.command == "" {
		return fmt.Errorf("--command is required")
	}
	if options.approvalID == "" {
		return fmt.Errorf("--approval-id is required before validation command execution")
	}
	return nil
}
