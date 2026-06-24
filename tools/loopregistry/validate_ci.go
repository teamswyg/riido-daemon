package main

import (
	"os"
	"strings"
)

func validateCIBindings(root string, reg registry) []string {
	var problems []string
	problems = append(problems, fileContains(root, reg.Workflow, reg.Command, "workflow")...)
	problems = append(problems, fileContains(root, ".pre-commit-config.yaml", reg.PrecommitHook, "pre-commit")...)
	problems = append(problems, fileContains(root, ".pre-commit-config.yaml", reg.Command, "pre-commit")...)
	problems = append(problems, validateWorkflowPathCoverage(root, reg)...)
	return problems
}

func fileContains(root, rel, token, owner string) []string {
	body, err := os.ReadFile(repoPath(root, rel))
	if err != nil {
		return []string{owner + " file missing: " + rel}
	}
	if strings.Contains(string(body), token) {
		return nil
	}
	return []string{owner + " must contain " + token}
}

func localFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
