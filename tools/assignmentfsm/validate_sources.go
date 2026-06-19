package main

import (
	"os"
	"strings"
)

func validateSources(repo string, checks []SourceCheck) ([]problem, []SourceCheckEvidence) {
	var problems []problem
	evidence := make([]SourceCheckEvidence, 0, len(checks))
	for _, check := range checks {
		path := repoPath(repo, check.File)
		body, err := os.ReadFile(path)
		ok := err == nil && strings.Contains(string(body), check.Contains)
		evidence = append(evidence, SourceCheckEvidence{Name: check.Name, File: check.File, OK: ok})
		if err != nil {
			problems = append(problems, problem{Message: check.Name + ": " + err.Error()})
			continue
		}
		if !ok {
			problems = append(problems, problem{Message: check.Name + ": missing expected source text"})
		}
	}
	return problems, evidence
}
