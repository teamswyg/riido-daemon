package main

import "fmt"

func validateStepReferences(manifest Manifest) []problem {
	known := make(map[string]bool, len(manifest.SourceChecks))
	for _, check := range manifest.SourceChecks {
		known[check.Name] = true
	}
	var problems []problem
	for _, step := range manifest.Steps {
		for _, name := range step.SourceChecks {
			if !known[name] {
				problems = append(problems, problem{fmt.Sprintf("unknown source check %q", name)})
			}
		}
	}
	return problems
}
