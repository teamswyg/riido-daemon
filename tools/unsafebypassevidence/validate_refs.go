package main

func validateRefs(m Manifest) []problem {
	known := map[string]bool{}
	for _, check := range m.SourceChecks {
		if known[check.Name] {
			return []problem{{Message: "duplicate source check " + check.Name}}
		}
		known[check.Name] = true
	}
	var problems []problem
	for _, surface := range m.Surfaces {
		for _, ref := range surface.SourceChecks {
			if !known[ref] {
				problems = append(problems, problem{Message: "unknown source check " + ref})
			}
		}
	}
	return problems
}
