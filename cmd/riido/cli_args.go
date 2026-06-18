package main

import (
	"errors"
	"fmt"
	"strconv"
)

var errCLIHelp = errors.New("cli help requested")

func isCLIHelp(err error) bool {
	return errors.Is(err, errCLIHelp)
}

func isHelpArg(arg string) bool {
	return arg == "--help" || arg == "-h"
}

func cliRequiredArg(args []string, index *int, flag, noun string) (string, error) {
	*index++
	if *index >= len(args) {
		return "", fmt.Errorf("%s requires a %s", flag, noun)
	}
	return args[*index], nil
}

func cliRequiredInt(args []string, index *int, flag string) (int, error) {
	value, err := cliRequiredArg(args, index, flag, "value")
	if err != nil {
		return 0, err
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", flag, err)
	}
	return parsed, nil
}

func cliRequiredBool(args []string, index *int, flag string) (bool, error) {
	value, err := cliRequiredArg(args, index, flag, "boolean")
	if err != nil {
		return false, err
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean: %w", flag, err)
	}
	return parsed, nil
}
