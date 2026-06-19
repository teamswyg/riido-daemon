package main

import (
	"fmt"
)

func validateAbsent(repo string, surfaces []AbsentSurface) ([]problem, []AbsentCheck) {
	var problems []problem
	checks := make([]AbsentCheck, 0, len(surfaces))
	for _, surface := range surfaces {
		check := absentCheck(repo, surface, &problems)
		if !check.Pass {
			problems = append(problems, problem{fmt.Sprintf("unexpected claim: %s", surface.Name)})
		}
		checks = append(checks, check)
	}
	return problems, checks
}

func absentCheck(repo string, surface AbsentSurface, problems *[]problem) AbsentCheck {
	check := AbsentCheck{Name: surface.Name, Scope: surface.Scope, Tokens: surface.Tokens, Hits: []string{}, Pass: true}
	for _, scope := range surface.Scope {
		scanAbsentScope(repo, scope, surface.Tokens, &check, problems)
	}
	check.Pass = len(check.Hits) == 0
	return check
}
