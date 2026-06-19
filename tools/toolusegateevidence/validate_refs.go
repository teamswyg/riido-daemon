package main

func validateRefs(m Manifest) []problem {
	known := map[string]bool{}
	for _, check := range m.SourceChecks {
		if check.Name == "" {
			return []problem{{Message: "source check with empty name"}}
		}
		known[check.Name] = true
	}
	var problems []problem
	for _, row := range appendActionFacts(m) {
		for _, ref := range row.SourceChecks {
			if !known[ref] {
				problems = append(problems, problem{Message: "unknown source check " + ref})
			}
		}
	}
	return problems
}

func appendActionFacts(m Manifest) []Fact {
	rows := append([]Fact{}, m.Facts...)
	for _, action := range m.ImplementedAction {
		rows = append(rows, Fact{Name: action.Name, SourceChecks: action.SourceChecks})
	}
	return rows
}
