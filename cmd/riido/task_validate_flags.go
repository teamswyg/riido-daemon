package main

import (
	"fmt"
	"time"
)

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
		err = parseTaskValidateTimeout(args, index, options)
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

func parseTaskValidateTimeout(args []string, index *int, options *taskValidateCLI) error {
	seconds, err := cliRequiredInt(args, index, "--timeout-seconds")
	if err != nil {
		return err
	}
	if seconds <= 0 {
		return fmt.Errorf("--timeout-seconds must be positive")
	}
	options.timeout = time.Duration(seconds) * time.Second
	return nil
}
