package main

import "fmt"

func validateCoverage(manifest Manifest, statuses map[string]string) ([]problem, []CoverageCheck) {
	covered := map[string]bool{}
	for _, row := range manifest.Mappings {
		covered[row.StatusConst] = true
	}
	var problems []problem
	checks := make([]CoverageCheck, 0, len(statuses))
	for statusConst, status := range statuses {
		check := CoverageCheck{StatusConst: statusConst, Status: status, Covered: covered[statusConst]}
		if !check.Covered {
			problems = append(problems, problem{fmt.Sprintf("unmapped ResultStatus const: %s", statusConst)})
		}
		checks = append(checks, check)
	}
	return problems, checks
}
