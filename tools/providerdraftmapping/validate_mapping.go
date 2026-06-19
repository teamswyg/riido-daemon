package main

func validateMappings(manifest Manifest, source map[string]string) ([]MappingCheck, []problem) {
	var checks []MappingCheck
	var problems []problem
	for _, row := range manifest.MappedEvents {
		actual := source[row.EventKind]
		passed := actual == row.EventTypeConst && eventTypeValues()[row.EventTypeConst] == row.EventType
		checks = append(checks, MappingCheck{EventKind: row.EventKind, Expected: row.EventTypeConst, Actual: actual, Passed: passed})
		if !passed {
			problems = append(problems, problem{"provider draft mapping drift: " + row.EventKind})
		}
	}
	for _, row := range manifest.SkippedEvents {
		actual := source[row.EventKind]
		passed := actual == ""
		checks = append(checks, MappingCheck{EventKind: row.EventKind, Expected: "skipped", Actual: actual, Passed: passed})
		if !passed {
			problems = append(problems, problem{"skipped event is mapped in source: " + row.EventKind})
		}
	}
	for eventKind := range source {
		if !manifestMentions(manifest, eventKind) {
			problems = append(problems, problem{"source mapping missing from manifest: " + eventKind})
		}
	}
	return checks, problems
}

func manifestMentions(manifest Manifest, eventKind string) bool {
	for _, row := range manifest.MappedEvents {
		if row.EventKind == eventKind {
			return true
		}
	}
	for _, row := range manifest.SkippedEvents {
		if row.EventKind == eventKind {
			return true
		}
	}
	return false
}
