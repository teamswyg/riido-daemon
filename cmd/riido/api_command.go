package main

import (
	"fmt"
)

func runAPI(args []string) error {
	if len(args) < 1 {
		printUsage()
		return fmt.Errorf("missing api command")
	}
	config, err := defaultAPICLIConfig()
	if err != nil {
		return err
	}

	switch apiCommand(args[0]) {
	case apiCommandStatus:
		return runAPIStatus(args[1:], config)
	case apiCommandTasks:
		return runAPITasks(args[1:], config)
	case apiCommandReviewDemo:
		return runAPIReviewDemo(args[1:], config)
	case apiCommandTransition:
		return runAPITransition(args[1:], config)
	case apiCommandEvidence:
		return runAPIEvidence(args[1:], config)
	case apiCommandValidate:
		return runAPIValidate(args[1:], config)
	default:
		printUsage()
		return fmt.Errorf("unknown api command: %s", args[0])
	}
}
