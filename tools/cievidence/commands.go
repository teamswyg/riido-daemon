package main

import "strings"

func workflowContainsCommand(text, command string) bool {
	return strings.Contains(text, command)
}
