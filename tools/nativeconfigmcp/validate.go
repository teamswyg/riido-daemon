package main

func validate(repo string, m Manifest) ([]problem, []SourceEvidence, []AbsentEvidence) {
	problems := validateRequired(m)
	problems = append(problems, validateFactRefs(m)...)
	sourceProblems, sources := validateSources(repo, m.SourceChecks)
	absentProblems, absent := validateAbsent(repo, m.AbsentSurfaces)
	problems = append(problems, sourceProblems...)
	problems = append(problems, absentProblems...)
	return problems, sources, absent
}

func validateRequired(m Manifest) []problem {
	var problems []problem
	required := map[string]string{
		"schema_version": m.SchemaVersion,
		"id":             m.ID,
		"title":          m.Title,
		"generated_doc":  m.GeneratedDoc,
		"workflow":       m.Workflow,
	}
	for name, value := range required {
		if value == "" {
			problems = append(problems, problem{Message: "missing " + name})
		}
	}
	if len(m.Facts) == 0 || len(m.SourceChecks) == 0 {
		problems = append(problems, problem{Message: "manifest needs facts and source checks"})
	}
	return problems
}
