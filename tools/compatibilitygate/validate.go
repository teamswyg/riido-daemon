package main

import "fmt"

func validate(repo string, manifest Manifest) ([]problem, []SourceResult) {
	var problems []problem
	problems = append(problems, validateRequired(manifest)...)
	problems = append(problems, validateReferences(manifest)...)
	sourceProblems, sources := validateSources(repo, manifest.SourceChecks)
	problems = append(problems, sourceProblems...)
	return problems, sources
}

func validateRequired(m Manifest) []problem {
	var problems []problem
	for _, value := range []string{m.SchemaVersion, m.ID, m.Title, m.GeneratedDoc, m.Workflow, m.EvidenceArtifact, m.Purpose} {
		if value == "" {
			problems = append(problems, problem{"manifest required field is empty"})
		}
	}
	if len(m.Inputs) == 0 || len(m.GateOrder) == 0 || len(m.SourceChecks) == 0 {
		problems = append(problems, problem{"manifest must define inputs, gate_order, and source_checks"})
	}
	return problems
}

func validateReferences(m Manifest) []problem {
	known := map[string]bool{}
	for _, check := range m.SourceChecks {
		if known[check.Name] {
			return []problem{{fmt.Sprintf("duplicate source check %q", check.Name)}}
		}
		known[check.Name] = true
	}
	return validateAllReferences(m, known)
}
