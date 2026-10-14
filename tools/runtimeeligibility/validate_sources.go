package main

import (
	"os"
	"strings"
)

type SourceEvidence struct {
	Name string `json:"name"`
	File string `json:"file"`
}

func validateSources(repo string, checks []SourceCheck) ([]problem, []SourceEvidence) {
	var problems []problem
	out := make([]SourceEvidence, 0, len(checks))
	seen := map[string]bool{}
	for _, check := range checks {
		if seen[check.Name] {
			problems = append(problems, problem{Message: "duplicate source check " + check.Name})
		}
		seen[check.Name] = true
		raw, err := os.ReadFile(repoPath(repo, check.File))
		if err != nil {
			problems = append(problems, problem{Message: err.Error()})
			continue
		}
		if !strings.Contains(string(raw), check.Contains) {
			problems = append(problems, problem{Message: check.File + " missing " + check.Name})
			continue
		}
		out = append(out, SourceEvidence{Name: check.Name, File: check.File})
	}
	return problems, out
}
