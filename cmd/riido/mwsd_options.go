package main

import "fmt"

type mwsdOptions struct {
	command    mwsdCommand
	socketPath string
	statePath  string
	taskDBPath string
	showUsage  bool
}

func parseMwsdOptions(args []string) (mwsdOptions, error) {
	options := mwsdOptions{command: mwsdCommand(args[0])}
	for index := 1; index < len(args); index++ {
		switch args[index] {
		case "--socket":
			value, next, err := mwsdFlagValue(args, index, "--socket")
			if err != nil {
				return options, err
			}
			options.socketPath = value
			index = next
		case "--state":
			value, next, err := mwsdFlagValue(args, index, "--state")
			if err != nil {
				return options, err
			}
			options.statePath = value
			index = next
		case "--task-db":
			value, next, err := mwsdFlagValue(args, index, "--task-db")
			if err != nil {
				return options, err
			}
			options.taskDBPath = value
			index = next
		case "--help", "-h":
			options.showUsage = true
			return options, nil
		default:
			return options, fmt.Errorf("unknown argument: %s", args[index])
		}
	}
	return options, nil
}

func mwsdFlagValue(args []string, index int, name string) (string, int, error) {
	next := index + 1
	if next >= len(args) {
		return "", index, fmt.Errorf("%s requires a path", name)
	}
	return args[next], next, nil
}

func mwsdHelpArg(arg string) bool {
	return arg == "--help" || arg == "-h"
}
