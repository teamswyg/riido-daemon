package main

func validateRefs(m Manifest) []problem {
	known := map[string]bool{}
	for _, check := range m.SourceChecks {
		known[check.Name] = true
	}
	var problems []problem
	for _, row := range m.Surfaces {
		problems = append(problems, validateCheckRefs(row.SourceChecks, known, row.Name)...)
	}
	for _, input := range m.Inputs {
		problems = append(problems, validateCheckRefs(input.SourceChecks, known, input.Name)...)
	}
	return problems
}

func validateCheckRefs(refs []string, known map[string]bool, owner string) []problem {
	var problems []problem
	if len(refs) == 0 {
		return []problem{{Message: owner + " has no source checks"}}
	}
	for _, ref := range refs {
		if !known[ref] {
			problems = append(problems, problem{Message: owner + " references unknown check " + ref})
		}
	}
	return problems
}
