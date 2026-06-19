package main

func validateCoverage(manifest Manifest) ([]CoverageCheck, []problem) {
	seen := map[string]string{}
	var checks []CoverageCheck
	var problems []problem
	add := func(kind, category string) {
		if previous, ok := seen[kind]; ok {
			problems = append(problems, problem{"duplicate event kind: " + kind + " in " + previous + " and " + category})
			return
		}
		seen[kind] = category
	}
	for _, row := range manifest.MappedEvents {
		if eventKindByConst()[row.EventKindConst] != row.EventKind {
			problems = append(problems, problem{"mapped event kind const mismatch: " + row.EventKind})
		}
		add(row.EventKind, "mapped")
	}
	for _, row := range manifest.SkippedEvents {
		if eventKindByConst()[row.EventKindConst] != row.EventKind {
			problems = append(problems, problem{"skipped event kind const mismatch: " + row.EventKind})
		}
		add(row.EventKind, "skipped")
	}
	for kind := range runtimeEventKinds() {
		category, ok := seen[kind]
		checks = append(checks, CoverageCheck{EventKind: kind, Category: category, Covered: ok})
		if !ok {
			problems = append(problems, problem{"runtime EventKind missing from manifest: " + kind})
		}
	}
	return checks, problems
}
