package main

func validate(repo string, m Manifest) ([]problem, []SourceResult, []ReservedRule) {
	var problems []problem
	problems = append(problems, validateRequired(m)...)
	problems = append(problems, validateReferences(m)...)
	sourceProblems, sources := validateSources(repo, m.SourceChecks)
	problems = append(problems, sourceProblems...)
	reserved := collectReserved(m)
	return problems, sources, reserved
}

func validateRequired(m Manifest) []problem {
	var problems []problem
	for _, value := range []string{m.SchemaVersion, m.ID, m.Title, m.GeneratedDoc, m.Workflow, m.EvidenceArtifact, m.Invariant} {
		if value == "" {
			problems = append(problems, problem{"manifest required field is empty"})
		}
	}
	if len(m.Inputs) == 0 || len(m.Flow) == 0 || len(m.SourceChecks) == 0 {
		problems = append(problems, problem{"manifest must define inputs, flow, and source_checks"})
	}
	return problems
}
