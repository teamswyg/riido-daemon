package main

import (
	"fmt"
	"time"

	"github.com/teamswyg/riido-daemon/internal/riidoapi"
	"github.com/teamswyg/riido-daemon/internal/validation"
)

func runAPIValidate(args []string, config apiCLIConfig) error {
	if len(args) < 1 {
		return fmt.Errorf("api validate requires a task id")
	}
	request := riidoapi.ValidateRequest{TaskID: args[0], Actor: "daemon", Source: "riido-api-cli"}
	for index := 1; index < len(args); index++ {
		if err := parseAPIValidateFlag(args, &index, &config, &request); err != nil {
			if isCLIHelp(err) {
				return nil
			}
			return err
		}
	}
	if request.Command == "" {
		return fmt.Errorf("--command is required")
	}
	if request.ApprovalID == "" {
		return fmt.Errorf("--approval-id is required before validation command execution")
	}
	timeout := validation.DefaultTimeout + 5*time.Second
	if request.TimeoutSeconds > 0 {
		timeout = time.Duration(request.TimeoutSeconds)*time.Second + 5*time.Second
	}
	var response riidoapi.ValidateResponse
	if err := requestAPI(config, timeout, "validate", request, &response); err != nil {
		return err
	}
	return printJSON(response)
}

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
		request.TimeoutSeconds, err = cliRequiredInt(args, index, "--timeout-seconds")
		if err == nil && request.TimeoutSeconds <= 0 {
			err = fmt.Errorf("--timeout-seconds must be positive")
		}
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
