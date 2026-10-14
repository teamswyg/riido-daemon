package main

import (
	"fmt"
	"os"
)

type claudePermissionMCPOptions struct {
	socket       string
	assignmentID string
	taskID       string
	runtimeID    string
}

func runClaudePermissionMCP(args []string) error {
	opts, err := parseClaudePermissionMCPOptions(args)
	if err != nil {
		return err
	}
	return serveClaudePermissionMCP(os.Stdin, os.Stdout, opts)
}

func parseClaudePermissionMCPOptions(args []string) (claudePermissionMCPOptions, error) {
	var opts claudePermissionMCPOptions
	var err error
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--socket":
			opts.socket, err = cliRequiredArg(args, &i, "--socket", "path")
		case "--assignment-id":
			opts.assignmentID, err = cliRequiredArg(args, &i, "--assignment-id", "id")
		case "--task-id":
			opts.taskID, err = cliRequiredArg(args, &i, "--task-id", "id")
		case "--runtime-id":
			opts.runtimeID, err = cliRequiredArg(args, &i, "--runtime-id", "id")
		default:
			return opts, fmt.Errorf("claude-permission-mcp: unknown arg %s", args[i])
		}
		if err != nil {
			return opts, err
		}
	}
	if opts.socket == "" || opts.assignmentID == "" || opts.taskID == "" || opts.runtimeID == "" {
		return opts, fmt.Errorf("claude-permission-mcp: --socket, --assignment-id, --task-id, and --runtime-id are required")
	}
	return opts, nil
}
