package main

import (
	"os"
	"strings"
)

func validateSources(repo string, checks []SourceCheck, problems *[]problem) []SourceCheckResult {
	results := make([]SourceCheckResult, 0, len(checks))
	for _, check := range checks {
		pass := sourceCheckPass(repo, check)
		results = append(results, SourceCheckResult{Name: check.Name, File: check.File, Pass: pass})
		if check.Name == "" || check.File == "" || check.Contains == "" {
			*problems = append(*problems, problem{"source checks require name, file, and contains"})
			continue
		}
		if !pass {
			*problems = append(*problems, problem{"source check failed: " + check.Name})
		}
	}
	return results
}

func sourceCheckPass(repo string, check SourceCheck) bool {
	data, err := os.ReadFile(repoPath(repo, check.File))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), check.Contains)
}
