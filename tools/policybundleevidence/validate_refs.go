package main

func validateRefs(m Manifest) []problem {
	sourceNames := map[string]bool{}
	for _, check := range m.SourceChecks {
		sourceNames[check.Name] = true
	}
	var problems []problem
	for _, fact := range m.Facts {
		if fact.Name == "" || fact.Summary == "" || len(fact.SourceChecks) == 0 {
			problems = append(problems, problem{"facts require name, summary, and source checks"})
		}
		for _, ref := range fact.SourceChecks {
			if !sourceNames[ref] {
				problems = append(problems, problem{"unknown source check: " + ref})
			}
		}
	}
	for _, boundary := range m.Boundaries {
		if boundary.Name == "" || boundary.Owner == "" || boundary.Summary == "" {
			problems = append(problems, problem{"boundaries require name, owner, and summary"})
		}
	}
	return problems
}
