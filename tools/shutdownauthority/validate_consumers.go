package main

import (
	"fmt"
	"strings"
)

func validateConsumers(repo string, manifest Manifest) ([]problem, []ConsumerCheck) {
	var problems []problem
	checks := make([]ConsumerCheck, 0, len(manifest.ConsumerRequirements))
	for _, row := range manifest.ConsumerRequirements {
		source, err := readSource(repo, row.File)
		check := ConsumerCheck{File: row.File, Contains: row.Contains, Reason: row.Reason}
		check.Pass = err == nil && strings.Contains(source, row.Contains)
		if !check.Pass {
			problems = append(problems, problem{fmt.Sprintf("lifecycle consumer drift: %s", row.File)})
		}
		checks = append(checks, check)
	}
	return problems, checks
}
