package main

import (
	"os"
	"strings"
)

func checkSources(repo string, manifest Manifest) []CheckResult {
	results := make([]CheckResult, 0, len(manifest.SourceChecks))
	for _, check := range manifest.SourceChecks {
		results = append(results, checkSource(repo, check))
	}
	return results
}

func checkSource(repo string, check SourceCheck) CheckResult {
	data, err := os.ReadFile(repoPath(repo, check.File))
	if err != nil {
		return CheckResult{Name: check.Name, File: check.File, Detail: err.Error()}
	}
	pass := strings.Contains(string(data), check.Contains)
	detail := ""
	if !pass {
		detail = "missing token"
	}
	return CheckResult{Name: check.Name, File: check.File, Pass: pass, Detail: detail}
}
