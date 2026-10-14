package main

import "time"

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

func defaultTaskValidateCLI(taskDBPath, taskID string) taskValidateCLI {
	return taskValidateCLI{
		taskDBPath: taskDBPath,
		taskID:     taskID,
		actor:      "daemon",
		source:     "riido-validation-runner",
	}
}
