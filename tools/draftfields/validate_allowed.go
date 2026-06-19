package main

import (
	"fmt"
	"strings"
)

func validateAllowed(repo string, fields []AllowedField) ([]problem, []AllowedCheck) {
	var problems []problem
	checks := make([]AllowedCheck, 0, len(fields))
	for _, field := range fields {
		check := allowedCheck(repo, field)
		if !check.Pass {
			problems = append(problems, problem{fmt.Sprintf("allowed field drift: %s", field.Field)})
		}
		checks = append(checks, check)
	}
	return problems, checks
}

func allowedCheck(repo string, field AllowedField) AllowedCheck {
	check := AllowedCheck{
		Field: field.Field, Status: field.Status,
		Source: field.Source, Contains: field.Contains,
	}
	if field.Status == "reserved" {
		check.Pass = field.Source == "" && field.Contains == ""
		return check
	}
	if field.Source == "" || field.Contains == "" {
		return check
	}
	source, err := readSource(repo, field.Source)
	check.Pass = err == nil && strings.Contains(source, field.Contains)
	return check
}
