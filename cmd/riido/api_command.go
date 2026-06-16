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
	method riidoapi.Method,
	request any,
	response any,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	client := riidoapi.NewClientWithTransport(config.transport, config.socketPath)
	client.Timeout = timeout
	return client.Request(ctx, string(method), request, response)
}
