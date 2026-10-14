package main

import "fmt"

type taskTransitionOptions struct {
	taskDBPath  string
	taskID      string
	toState     string
	eventType   string
	actor       string
	source      string
	reason      string
	provider    string
	decisionLLM string
	approvalID  string
	commandID   string
}

func parseTaskTransitionOptions(args []string, taskDBPath string) (taskTransitionOptions, error) {
	opts := taskTransitionOptions{taskDBPath: taskDBPath, taskID: args[0], actor: "human", source: "riido-cli"}
	for index := 1; index < len(args); index++ {
		if err := parseTaskTransitionFlag(args, &index, &opts); err != nil {
			return opts, err
		}
	}
	return opts, nil
}

func parseTaskTransitionFlag(args []string, index *int, opts *taskTransitionOptions) error {
	var err error
	switch args[*index] {
	case "--task-db":
		opts.taskDBPath, err = cliRequiredArg(args, index, "--task-db", "path")
	case "--to":
		opts.toState, err = cliRequiredArg(args, index, "--to", "state")
	case "--event":
		opts.eventType, err = cliRequiredArg(args, index, "--event", "event type")
	case "--actor":
		opts.actor, err = cliRequiredArg(args, index, "--actor", "value")
	case "--source":
		opts.source, err = cliRequiredArg(args, index, "--source", "value")
	case "--reason":
		opts.reason, err = cliRequiredArg(args, index, "--reason", "value")
	case "--provider":
		opts.provider, err = cliRequiredArg(args, index, "--provider", "value")
	case "--decision-llm":
		opts.decisionLLM, err = cliRequiredArg(args, index, "--decision-llm", "value")
	case "--approval-id":
		opts.approvalID, err = cliRequiredArg(args, index, "--approval-id", "value")
	case "--command-id":
		opts.commandID, err = cliRequiredArg(args, index, "--command-id", "value")
	case "--help", "-h":
		printUsage()
		return errCLIHelp
	default:
		return fmt.Errorf("unknown argument: %s", args[*index])
	}
	return err
}
