package main

import (
	"fmt"
	"time"

	"github.com/teamswyg/riido-daemon/internal/riidoapi"
)

func runAPITransition(args []string, config apiCLIConfig) error {
	if len(args) < 1 {
		return fmt.Errorf("api transition requires a task id")
	}
	request := riidoapi.TransitionRequest{TaskID: args[0], Actor: "human", Source: "riido-api-cli"}
	for index := 1; index < len(args); index++ {
		if err := parseAPITransitionFlag(args, &index, &config, &request); err != nil {
			if isCLIHelp(err) {
				return nil
			}
			return err
		}
	}
	if request.ToState == "" {
		return fmt.Errorf("--to is required")
	}
	if request.EventType == "" {
		return fmt.Errorf("--event is required")
	}
	var response riidoapi.TransitionResponse
	if err := requestAPI(config, 5*time.Second, "transition", request, &response); err != nil {
		return err
	}
	return printJSON(response)
}

func parseAPITransitionFlag(
	args []string,
	index *int,
	config *apiCLIConfig,
	request *riidoapi.TransitionRequest,
) error {
	if handled, err := parseAPIConnectionFlag(args, index, config); handled || err != nil {
		return err
	}
	var err error
	switch args[*index] {
	case "--to":
		request.ToState, err = cliRequiredArg(args, index, "--to", "state")
	case "--event":
		request.EventType, err = cliRequiredArg(args, index, "--event", "event type")
	case "--actor":
		request.Actor, err = cliRequiredArg(args, index, "--actor", "value")
	case "--source":
		request.Source, err = cliRequiredArg(args, index, "--source", "value")
	case "--reason":
		request.Reason, err = cliRequiredArg(args, index, "--reason", "value")
	case "--provider":
		request.Provider, err = cliRequiredArg(args, index, "--provider", "value")
	case "--decision-llm":
		request.DecisionLLM, err = cliRequiredArg(args, index, "--decision-llm", "value")
	case "--approval-id":
		request.ApprovalID, err = cliRequiredArg(args, index, "--approval-id", "value")
	case "--command-id":
		request.CommandID, err = cliRequiredArg(args, index, "--command-id", "value")
	default:
		return fmt.Errorf("unknown argument: %s", args[*index])
	}
	return err
}
