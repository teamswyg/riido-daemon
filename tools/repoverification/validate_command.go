package main

import "fmt"

func validateCommand(command commandSpec, seen map[string]bool) []string {
	var problems []string
	if command.ID == "" || command.Description == "" || len(command.Argv) == 0 {
		problems = append(problems, "commands require id, description, and argv")
	}
	if seen[command.ID] {
		problems = append(problems, fmt.Sprintf("duplicate command id %q", command.ID))
	}
	seen[command.ID] = true
	if len(command.Argv) > 0 && command.Argv[0] == "" {
		problems = append(problems, fmt.Sprintf("%s argv[0] must not be empty", command.ID))
	}
	return problems
}

func anyFailed(commands []commandEvidence) bool {
	for _, command := range commands {
		if command.Status == "failed" {
			return true
		}
	}
	return false
}
