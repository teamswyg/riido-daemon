package main

import (
	"os"
	"strings"
)

func checkSources(repo string, manifest Manifest) []CheckResult {
	results := make([]CheckResult, 0, len(manifest.SourceChecks))
	for _, check := range manifest.SourceChecks {
		results = append(results, checkSource(repo, check.Name, check.File, check.Contains))
	}
	return results
}

func checkAnchors(repo string, manifest Manifest) []CheckResult {
	results := make([]CheckResult, 0, len(manifest.CoverageAnchors))
	for _, anchor := range manifest.CoverageAnchors {
		results = append(results, checkSource(repo, "anchor-"+anchor.Name, anchor.File, anchor.Contains))
	}
	return results
}

func checkSource(repo, name, file, want string) CheckResult {
	data, err := os.ReadFile(repoPath(repo, file))
	if err != nil {
		return CheckResult{Name: name, File: file, Detail: err.Error()}
	}
	pass := strings.Contains(string(data), want)
	detail := ""
	if !pass {
		detail = "missing token"
	}
	return CheckResult{Name: name, File: file, Pass: pass, Detail: detail}
}
