package main

import "fmt"

func validateMappings(
	manifest Manifest,
	statuses map[string]string,
	source terminalSource,
	events map[string]string,
) ([]problem, []MappingCheck) {
	var problems []problem
	checks := make([]MappingCheck, 0, len(manifest.Mappings))
	for _, row := range manifest.Mappings {
		actual, resolution := actualEvent(row.StatusConst, source)
		check := MappingCheck{
			StatusConst: row.StatusConst, Status: row.Status,
			ExpectedEventConst: row.EventTypeConst, ActualEventConst: actual,
			ActualResolution: resolution, ExpectedEventValue: row.EventType,
			ContractEventValue: events[row.EventTypeConst],
		}
		check.Pass = statuses[row.StatusConst] == row.Status &&
			actual == row.EventTypeConst && events[row.EventTypeConst] == row.EventType
		if !check.Pass {
			problems = append(problems, problem{fmt.Sprintf("mapping drift for %s", row.StatusConst)})
		}
		checks = append(checks, check)
	}
	return problems, checks
}

func actualEvent(statusConst string, source terminalSource) (string, string) {
	if eventType := source.Cases[statusConst]; eventType != "" {
		return eventType, "explicit"
	}
	return source.Fallback, "fallback"
}
