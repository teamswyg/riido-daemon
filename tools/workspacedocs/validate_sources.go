package main

import (
	"path/filepath"
	"strings"
)

func validateSourceChecks(repo string, checks []sourceCheck) ([]sourceCheckResult, []string) {
	var results []sourceCheckResult
	var problems []string
	for _, check := range checks {
		result := sourceCheckResult{Name: check.Name, File: check.File}
		text, err := readString(filepath.Join(repo, check.File))
		if err == nil && strings.Contains(text, check.Contains) {
			result.Pass = true
		} else {
			problems = append(problems, "source check failed: "+check.Name)
		}
		results = append(results, result)
	}
	return results, problems
}

func mustExist(repo string, paths ...string) []string {
	var problems []string
	for _, path := range paths {
		if _, err := readString(filepath.Join(repo, path)); err != nil {
			problems = append(problems, "missing required file: "+path)
		}
	}
	return problems
}
