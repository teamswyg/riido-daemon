package main

import "fmt"

func validateDefaults(manifest Manifest, source terminalSource, events map[string]string) ([]problem, []DefaultCheck) {
	checks := []DefaultCheck{
		{
			Name: "empty_status", Expected: manifest.Defaults.EmptyStatusConst,
			Actual: source.EmptyDefault,
		},
		{
			Name: "fallback_event_type", Expected: manifest.Defaults.FallbackEventTypeConst,
			Actual: source.Fallback,
		},
		{
			Name: "fallback_event_value", Expected: manifest.Defaults.FallbackEventType,
			Actual: events[manifest.Defaults.FallbackEventTypeConst],
		},
	}
	var problems []problem
	for i := range checks {
		checks[i].Pass = checks[i].Expected == checks[i].Actual
		if !checks[i].Pass {
			problems = append(problems, problem{fmt.Sprintf("default drift for %s", checks[i].Name)})
		}
	}
	return problems, checks
}
