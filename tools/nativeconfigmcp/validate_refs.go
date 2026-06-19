package main

func validateFactRefs(m Manifest) []problem {
	known := map[string]bool{}
	for _, check := range m.SourceChecks {
		if check.Name == "" {
			return []problem{{Message: "source check with empty name"}}
		}
		known[check.Name] = true
	}
	var problems []problem
	for _, fact := range m.Facts {
		if fact.Name == "" {
			problems = append(problems, problem{Message: "fact with empty name"})
		}
		for _, ref := range fact.SourceChecks {
			if !known[ref] {
				problems = append(problems, problem{Message: "unknown source check " + ref})
			}
		}
	}
	return problems
}
