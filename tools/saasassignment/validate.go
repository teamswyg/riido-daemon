package main

func validate(repo string, m Manifest) ([]problem, []SourceResult, []AbsentCheck) {
	var problems []problem
	problems = append(problems, validateRequired(m)...)
	problems = append(problems, validateFactReferences(m)...)
	sourceProblems, sources := validateSources(repo, m.SourceChecks)
	absentProblems, absent := validateAbsent(repo, m.AbsentSurfaces)
	problems = append(problems, sourceProblems...)
	problems = append(problems, absentProblems...)
	return problems, sources, absent
}

func validateRequired(m Manifest) []problem {
	var problems []problem
	for _, value := range []string{
		m.SchemaVersion, m.ID, m.Title, m.GeneratedDoc, m.MigrationDoc,
		m.Workflow, m.EvidenceArtifact, m.Purpose,
	} {
		if value == "" {
			problems = append(problems, problem{"manifest required field is empty"})
		}
	}
	if len(m.Facts) == 0 || len(m.SourceChecks) == 0 || len(m.Assertions) == 0 {
		problems = append(problems, problem{"manifest must define facts, source_checks, and assertions"})
	}
	return problems
}
