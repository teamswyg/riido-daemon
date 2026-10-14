package main

import (
	"fmt"
	"slices"
)

func validateAWSOperations(m manifest) []string {
	var problems []string
	if len(m.AllowedAWSOperations) == 0 {
		problems = append(problems, "allowed_aws_operations must not be empty")
	}
	for _, op := range m.ForbiddenAWSOps {
		if contains(m.AllowedAWSOperations, op) {
			problems = append(problems, fmt.Sprintf("forbidden AWS operation %q is also allowed", op))
		}
	}
	problems = append(problems, requireAll("forbidden_aws_operations", m.ForbiddenAWSOps, requiredForbiddenOps)...)
	if !contains(m.AllowedAWSOperations, "ssm:DescribeParameters") {
		problems = append(problems, "allowed_aws_operations must include ssm:DescribeParameters")
	}
	return problems
}

func requireAll(label string, got, required []string) []string {
	var problems []string
	for _, value := range required {
		if !contains(got, value) {
			problems = append(problems, fmt.Sprintf("%s missing %q", label, value))
		}
	}
	return problems
}

func contains(values []string, target string) bool {
	return slices.Contains(values, target)
}
