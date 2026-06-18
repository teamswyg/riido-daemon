package main

import (
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/riidoapi"
)

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
		return parseAPISocketFlag(args, index, config)
	case "--transport":
		return parseAPITransportFlag(args, index, config)
	case "--help", "-h":
		printUsage()
		return true, errCLIHelp
	default:
		return false, nil
	}
}

func parseAPISocketFlag(args []string, index *int, config *apiCLIConfig) (bool, error) {
	value, err := cliRequiredArg(args, index, "--socket", "path")
	if err != nil {
		return true, err
	}
	config.socketPath = value
	return true, nil
}

func parseAPITransportFlag(args []string, index *int, config *apiCLIConfig) (bool, error) {
	value, err := cliRequiredArg(args, index, "--transport", "value")
	if err != nil {
		return true, err
	}
	config.transport = riidoapi.LocalTransport(value)
	return true, nil
}
