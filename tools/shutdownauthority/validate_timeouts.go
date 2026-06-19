package main

import "fmt"

func validateTimeouts(manifest Manifest, source map[string]string) ([]problem, []TimeoutCheck) {
	var problems []problem
	checks := make([]TimeoutCheck, 0, len(manifest.Timeouts))
	for _, row := range manifest.Timeouts {
		check := TimeoutCheck{
			Const: row.Const, ExpectedDuration: row.Duration,
			ActualDuration: source[row.Const],
		}
		check.Pass = check.ExpectedDuration == check.ActualDuration
		if !check.Pass {
			problems = append(problems, problem{fmt.Sprintf("shutdown timeout drift: %s", row.Const)})
		}
		checks = append(checks, check)
	}
	return problems, checks
}
