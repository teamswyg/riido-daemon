package main

import (
	"strings"
)

func checkHelpOutput(repo string, manifest Manifest) CheckResult {
	out, err := runGoCommand(repo, "run", "./cmd/riido", "--help")
	result := CheckResult{Name: "help-output", Command: "go run ./cmd/riido --help", Pass: err == nil}
	if err != nil {
		result.Detail = err.Error()
		return result
	}
	for _, group := range manifest.CommandGroups {
		for _, token := range group.UsageTokens {
			if !strings.Contains(out, token) {
				result.Pass = false
				result.Detail = "missing usage token: " + token
				return result
			}
		}
	}
	return result
}
