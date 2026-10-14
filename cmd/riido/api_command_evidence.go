package main

import (
	"fmt"
	"time"

	"github.com/teamswyg/riido-daemon/internal/riidoapi"
)

func runAPIEvidence(args []string, config apiCLIConfig) error {
	if len(args) < 1 {
		return fmt.Errorf("api evidence requires a task id")
	}
	request := riidoapi.EvidenceRequest{TaskID: args[0], Actor: "daemon", Source: "riido-api-cli"}
	for index := 1; index < len(args); index++ {
		if err := parseAPIEvidenceFlag(args, &index, &config, &request); err != nil {
			if isCLIHelp(err) {
				return nil
			}
			return err
		}
	}
	if request.Command == "" {
		return fmt.Errorf("--command is required")
	}
	var response riidoapi.EvidenceResponse
	if err := requestAPI(config, 5*time.Second, riidoapi.MethodEvidence, request, &response); err != nil {
		return err
	}
	return printJSON(response)
}

func parseAPIEvidenceFlag(args []string, index *int, config *apiCLIConfig, request *riidoapi.EvidenceRequest) error {
	if handled, err := parseAPIConnectionFlag(args, index, config); handled || err != nil {
		return err
	}
	var err error
	switch args[*index] {
	case "--command":
		request.Command, err = cliRequiredArg(args, index, "--command", "value")
	case "--exit-code":
		request.ExitCode, err = cliRequiredInt(args, index, "--exit-code")
	case "--result":
		request.Result, err = cliRequiredArg(args, index, "--result", "value")
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
	case "--provider-run-id":
		request.ProviderRunID, err = cliRequiredArg(args, index, "--provider-run-id", "value")
	case "--provider-run-result":
		request.ProviderRunResult, err = cliRequiredArg(args, index, "--provider-run-result", "value")
	default:
		return fmt.Errorf("unknown argument: %s", args[*index])
	}
	return err
}
