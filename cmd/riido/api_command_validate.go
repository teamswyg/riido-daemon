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
	if err := requestAPI(config, timeout, riidoapi.MethodValidate, request, &response); err != nil {
		return err
	}
	return printJSON(response)
}
