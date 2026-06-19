package main

import (
	"fmt"
	"os"
	"strings"
)

func validateSourceChecks(repo string, checks []sourceCheck) ([]sourceCheckResult, []string) {
	var results []sourceCheckResult
	var problems []string
	for _, check := range checks {
		passed, err := sourceCheckPasses(repo, check)
		results = append(results, sourceCheckResult{Name: check.Name, File: check.File, Passed: passed})
		if err != nil {
			problems = append(problems, err.Error())
		}
	}
	return results, problems
}

func sourceCheckPasses(repo string, check sourceCheck) (bool, error) {
	if check.Name == "" || check.File == "" || check.Contains == "" {
		return false, fmt.Errorf("source check requires name, file, and contains")
	}
	data, err := os.ReadFile(repoPath(repo, check.File))
	if err != nil {
		return false, err
	}
	if !strings.Contains(string(data), check.Contains) {
		return false, fmt.Errorf("source check %q missing anchor in %s", check.Name, check.File)
	}
	return true, nil
}

func mustExist(repo, rel string) []string {
	if _, err := os.Stat(repoPath(repo, rel)); err != nil {
		return []string{fmt.Sprintf("missing artifact %q", rel)}
	}
	return nil
}
