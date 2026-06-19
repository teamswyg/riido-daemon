package main

import (
	"os"
	"strings"
)

func checkSources(repo string, checks []sourceCheck) []checkResult {
	results := make([]checkResult, 0, len(checks))
	for _, check := range checks {
		results = append(results, checkSource(repo, check))
	}
	return results
}

func checkSource(repo string, check sourceCheck) checkResult {
	data, err := os.ReadFile(repoPath(repo, check.File))
	if err != nil {
		return checkResult{Name: check.Name, File: check.File, Detail: err.Error()}
	}
	pass := strings.Contains(string(data), check.Contains)
	detail := ""
	if !pass {
		detail = "missing token"
	}
	return checkResult{Name: check.Name, File: check.File, Pass: pass, Detail: detail}
}
