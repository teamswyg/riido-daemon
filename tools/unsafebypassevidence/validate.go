package main

func validate(repo string, m Manifest) (
	[]problem,
	[]SourceEvidence,
	[]PolicyEvidence,
	[]CodexArgEvidence,
) {
	problems := validateRequired(m)
	problems = append(problems, validateRefs(m)...)
	sourceProblems, sources := validateSources(repo, m.SourceChecks)
	policyProblems, policies := validatePolicy(m.Surfaces)
	codexProblems, codexArgs := validateCodexArgs(m.Surfaces)
	problems = append(problems, sourceProblems...)
	problems = append(problems, policyProblems...)
	problems = append(problems, codexProblems...)
	return problems, sources, policies, codexArgs
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
	if len(m.Surfaces) == 0 || len(m.Assertions) == 0 {
		problems = append(problems, problem{Message: "manifest needs surfaces and assertions"})
	}
	return problems
}
