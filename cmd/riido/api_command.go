package main

import (
	"context"
	"fmt"
	"time"

	"github.com/teamswyg/riido-daemon/internal/riidoapi"
)

type apiCLIConfig struct {
	socketPath string
	transport  riidoapi.LocalTransport
}

func runAPI(args []string) error {
	if len(args) < 1 {
		printUsage()
		return fmt.Errorf("missing api command")
	}
	config, err := defaultAPICLIConfig()
	if err != nil {
		return err
	}

	switch args[0] {
	case "status":
		return runAPIStatus(args[1:], config)
	case "tasks":
		return runAPITasks(args[1:], config)
	case "review-demo":
		return runAPIReviewDemo(args[1:], config)
	case "transition":
		return runAPITransition(args[1:], config)
	case "evidence":
		return runAPIEvidence(args[1:], config)
	case "validate":
		return runAPIValidate(args[1:], config)
	default:
		printUsage()
		return fmt.Errorf("unknown api command: %s", args[0])
	}
}

func defaultAPICLIConfig() (apiCLIConfig, error) {
	socketPath, err := riidoapi.DefaultSocketPath()
	if err != nil {
		return apiCLIConfig{}, err
	}
	return apiCLIConfig{socketPath: socketPath, transport: riidoapi.LocalTransportUnixSocket}, nil
}

func parseAPIConnectionArgs(args []string, config apiCLIConfig) (apiCLIConfig, error) {
	for index := 0; index < len(args); index++ {
		handled, err := parseAPIConnectionFlag(args, &index, &config)
		if err != nil {
			return apiCLIConfig{}, err
		}
		if !handled {
			return apiCLIConfig{}, fmt.Errorf("unknown argument: %s", args[index])
		}
	}
	return config, nil
}

func parseAPIConnectionFlag(args []string, index *int, config *apiCLIConfig) (bool, error) {
	switch args[*index] {
	case "--socket":
		value, err := cliRequiredArg(args, index, "--socket", "path")
		if err != nil {
			return true, err
		}
		config.socketPath = value
	case "--transport":
		value, err := cliRequiredArg(args, index, "--transport", "value")
		if err != nil {
			return true, err
		}
		config.transport = riidoapi.LocalTransport(value)
	case "--help", "-h":
		printUsage()
		return true, errCLIHelp
	default:
		return false, nil
	}
	return true, nil
}

func requestAPI(
	config apiCLIConfig,
	timeout time.Duration,
	method string,
	request any,
	response any,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	client := riidoapi.NewClientWithTransport(config.transport, config.socketPath)
	client.Timeout = timeout
	return client.Request(ctx, method, request, response)
}
