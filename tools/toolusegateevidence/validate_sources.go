package main

import (
	"os"
	"strings"
)

func validateSources(repo string, checks []SourceCheck) ([]problem, []SourceEvidence) {
	var problems []problem
	evidence := make([]SourceEvidence, 0, len(checks))
	for _, check := range checks {
		body, err := os.ReadFile(repoPath(repo, check.File))
		ok := err == nil && strings.Contains(string(body), check.Contains)
		evidence = append(evidence, SourceEvidence{Name: check.Name, File: check.File, OK: ok})
		if err != nil {
			problems = append(problems, problem{Message: check.Name + ": " + err.Error()})
			continue
		}
		if check.Contains == "" || !ok {
			problems = append(problems, problem{Message: "missing source evidence " + check.Name})
		}
	}
	return problems, evidence
}
