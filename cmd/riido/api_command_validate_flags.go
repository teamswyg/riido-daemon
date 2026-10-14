package main

import (
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/riidoapi"
)

func parseAPIValidateFlag(args []string, index *int, config *apiCLIConfig, request *riidoapi.ValidateRequest) error {
	if handled, err := parseAPIConnectionFlag(args, index, config); handled || err != nil {
		return err
	}
	var err error
	switch args[*index] {
	case "--command":
		request.Command, err = cliRequiredArg(args, index, "--command", "value")
	case "--workdir":
		request.Workdir, err = cliRequiredArg(args, index, "--workdir", "path")
	case "--timeout-seconds":
		err = parseAPIValidateTimeoutFlag(args, index, request)
	case "--actor":
		request.Actor, err = cliRequiredArg(args, index, "--actor", "value")
	case "--source":
		request.Source, err = cliRequiredArg(args, index, "--source", "value")
	case "--summary":
		request.Summary, err = cliRequiredArg(args, index, "--summary", "value")
	case "--provider":
		request.Provider, err = cliRequiredArg(args, index, "--provider", "value")
	case "--decision-llm":
		request.DecisionLLM, err = cliRequiredArg(args, index, "--decision-llm", "value")
	case "--approval-id":
		request.ApprovalID, err = cliRequiredArg(args, index, "--approval-id", "value")
	case "--command-id":
		request.CommandID, err = cliRequiredArg(args, index, "--command-id", "value")
	case "--validation-gate":
		request.ValidationGate, err = cliRequiredArg(args, index, "--validation-gate", "value")
	default:
		return fmt.Errorf("unknown argument: %s", args[*index])
	}
	return err
}

func parseAPIValidateTimeoutFlag(args []string, index *int, request *riidoapi.ValidateRequest) error {
	timeout, err := cliRequiredInt(args, index, "--timeout-seconds")
	if err != nil {
		return err
	}
	if timeout <= 0 {
		return fmt.Errorf("--timeout-seconds must be positive")
	}
	request.TimeoutSeconds = timeout
	return nil
}
