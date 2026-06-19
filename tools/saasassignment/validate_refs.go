package main

import "fmt"

func validateFactReferences(m Manifest) []problem {
	known, problems := knownSourceChecks(m.SourceChecks)
	for _, fact := range m.Facts {
		if fact.Status == "implemented" && len(fact.SourceChecks) == 0 {
			problems = append(problems, problem{fmt.Sprintf("implemented fact %q has no source checks", fact.Name)})
		}
		for _, ref := range fact.SourceChecks {
			if !known[ref] {
				problems = append(problems, problem{fmt.Sprintf("fact %q references unknown source check %q", fact.Name, ref)})
			}
		}
	}
	return problems
}

func knownSourceChecks(checks []SourceCheck) (map[string]bool, []problem) {
	known := map[string]bool{}
	var problems []problem
	for _, check := range checks {
		if known[check.Name] {
			problems = append(problems, problem{fmt.Sprintf("duplicate source check %q", check.Name)})
		}
		known[check.Name] = true
	}
	return known, problems
}
