package main

func validate(repo string, manifest Manifest) ([]problem, []MappingCheck, []CoverageCheck, []DefaultCheck) {
	statuses, err := resultStatusValues(repo, manifest)
	if err != nil {
		return []problem{{err.Error()}}, nil, nil, nil
	}
	source, err := terminalSourceMapping(repo, manifest)
	if err != nil {
		return []problem{{err.Error()}}, nil, nil, nil
	}
	events := eventTypeValues()
	mappingProblems, mappings := validateMappings(manifest, statuses, source, events)
	coverageProblems, coverage := validateCoverage(manifest, statuses)
	defaultProblems, defaults := validateDefaults(manifest, source, events)
	problems := mappingProblems
	problems = append(problems, coverageProblems...)
	problems = append(problems, defaultProblems...)
	return problems, mappings, coverage, defaults
}
