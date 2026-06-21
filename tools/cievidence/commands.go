package main

import "strings"

func requiredCommandsFor(workflow string) []string {
	if strings.HasSuffix(workflow, "go-ci.yml") {
		return []string{
			"go mod download",
			"go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2",
			"golangci-lint run ./... --timeout=5m",
			"go test ./...",
		}
	}
	return []string{
		"go list -m all",
		"go test ./...",
	}
}

func workflowContainsCommand(text, command string) bool {
	return strings.Contains(text, command)
}
