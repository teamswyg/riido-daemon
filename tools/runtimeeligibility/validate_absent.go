package main

import (
	"os"
	"strings"
)

type AbsentEvidence struct {
	Name   string   `json:"name"`
	Scope  []string `json:"scope"`
	Tokens []string `json:"tokens"`
}

func validateAbsent(repo string, scans []AbsentScan) ([]problem, []AbsentEvidence) {
	var problems []problem
	out := make([]AbsentEvidence, 0, len(scans))
	for _, scan := range scans {
		scanProblems := validateOneAbsent(repo, scan)
		problems = append(problems, scanProblems...)
		out = append(out, AbsentEvidence{Name: scan.Name, Scope: scan.Scope, Tokens: scan.Tokens})
	}
	return problems, out
}

func validateOneAbsent(repo string, scan AbsentScan) []problem {
	var problems []problem
	for _, rel := range scan.Scope {
		raw, err := os.ReadFile(repoPath(repo, rel))
		if err != nil {
			problems = append(problems, problem{Message: err.Error()})
			continue
		}
		for _, token := range scan.Tokens {
			if strings.Contains(string(raw), token) {
				problems = append(problems, problem{Message: rel + " contains forbidden token " + token})
			}
		}
	}
	return problems
}
