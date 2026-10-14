package main

import "fmt"

func validateAllReferences(m Manifest, known map[string]bool) []problem {
	var problems []problem
	for _, input := range m.Inputs {
		problems = append(problems, validateRefSet("input", input.Name, input.SourceChecks, known)...)
	}
	for _, step := range m.GateOrder {
		problems = append(problems, validateRefSet("step", step.Step, step.SourceChecks, known)...)
	}
	for _, failure := range m.FailureSemantics {
		problems = append(problems, validateRefSet("failure", failure.Case, failure.SourceChecks, known)...)
	}
	return problems
}

func validateRefSet(kind, name string, refs []string, known map[string]bool) []problem {
	if len(refs) == 0 {
		return []problem{{fmt.Sprintf("%s %q has no source checks", kind, name)}}
	}
	var problems []problem
	for _, ref := range refs {
		if !known[ref] {
			problems = append(problems, problem{fmt.Sprintf("%s %q references unknown source check %q", kind, name, ref)})
		}
	}
	return problems
}
