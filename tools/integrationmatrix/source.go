package main

import (
	"os"
	"strings"
)

type checkResult struct {
	Name   string `json:"name"`
	File   string `json:"file"`
	Pass   bool   `json:"pass"`
	Detail string `json:"detail,omitempty"`
}

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
	return checkResult{Name: check.Name, File: check.File, Pass: pass, Detail: detailFor(pass)}
}

func detailFor(pass bool) string {
	if pass {
		return ""
	}
	return "missing token"
}
