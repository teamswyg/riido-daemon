package main

func validateRefs(manifest Manifest) []problem {
	known := map[string]bool{}
	for _, check := range manifest.SourceChecks {
		if known[check.Name] {
			return []problem{{Message: "duplicate source check " + check.Name}}
		}
		known[check.Name] = true
	}
	var problems []problem
	for _, fact := range manifest.Facts {
		for _, ref := range fact.SourceChecks {
			if !known[ref] {
				problems = append(problems, problem{Message: "unknown source check " + ref})
			}
		}
	}
	return problems
}
