package main

import (
	"fmt"
	"reflect"
)

func evaluate(item requiredEvidence, data map[string]any) ([]checkSummary, []string) {
	var checks []checkSummary
	var problems []string
	for _, assertion := range item.Assertions {
		status := statusVerified
		value, ok := data[assertion.Field]
		if !ok {
			status = statusFailed
			problems = append(problems, item.ID+" missing "+assertion.Field)
		} else if !assertionPasses(assertion, value) {
			status = statusFailed
			problems = append(problems, failedMessage(item.ID, assertion, value))
		}
		checks = append(checks, checkSummary{EvidenceID: item.ID, Field: assertion.Field, Status: status})
	}
	return checks, problems
}

func assertionPasses(assertion assertion, value any) bool {
	if assertion.Empty {
		return isEmpty(value)
	}
	return reflect.DeepEqual(value, assertion.Equals)
}

func isEmpty(value any) bool {
	switch typed := value.(type) {
	case []any:
		return len(typed) == 0
	case map[string]any:
		return len(typed) == 0
	case string:
		return typed == ""
	default:
		return value == nil
	}
}

func failedMessage(id string, assertion assertion, value any) string {
	if assertion.Empty {
		return fmt.Sprintf("%s %s is not empty", id, assertion.Field)
	}
	return fmt.Sprintf("%s %s = %v, want %v", id, assertion.Field, value, assertion.Equals)
}
