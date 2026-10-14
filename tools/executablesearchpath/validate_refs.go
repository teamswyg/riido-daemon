package main

func validateRefs(m Manifest) []problem {
	known := map[string]bool{}
	for _, check := range m.SourceChecks {
		known[check.Name] = true
	}
	var problems []problem
	for _, row := range m.SearchOrder {
		problems = append(problems, validateCheckRefs(row.Name, row.SourceChecks, known)...)
	}
	for _, row := range m.Rules {
		problems = append(problems, validateCheckRefs(row.Name, row.SourceChecks, known)...)
	}
	return problems
}

func validateCheckRefs(owner string, refs []string, known map[string]bool) []problem {
	if len(refs) == 0 {
		return []problem{{Message: owner + " has no source checks"}}
	}
	var problems []problem
	for _, ref := range refs {
		if !known[ref] {
			problems = append(problems, problem{Message: owner + " references unknown check " + ref})
		}
	}
	return problems
}
