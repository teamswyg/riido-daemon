package main

func validate(repo string, m Manifest) (
	[]problem,
	[]SourceEvidence,
	[]SurfaceEvidence,
) {
	problems := validateRequired(m)
	sourceProblems, sourceEvidence := validateSources(repo, m.SourceChecks)
	surfaceProblems, surfaceEvidence := validateSurfaces(m.Surfaces)
	problems = append(problems, validateRefs(m)...)
	problems = append(problems, sourceProblems...)
	problems = append(problems, surfaceProblems...)
	return problems, sourceEvidence, surfaceEvidence
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
	if len(m.Surfaces) == 0 || len(m.SourceChecks) == 0 {
		problems = append(problems, problem{Message: "manifest needs surfaces and source checks"})
	}
	if len(m.Inputs) == 0 || len(m.Assertions) == 0 {
		problems = append(problems, problem{Message: "manifest needs inputs and assertions"})
	}
	return problems
}
