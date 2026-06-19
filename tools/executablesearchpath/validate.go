package main

func validate(repo string, m Manifest) (
	[]problem,
	[]SourceEvidence,
	[]BehaviorEvidence,
) {
	problems := validateRequired(m)
	sourceProblems, sourceEvidence := validateSources(repo, m.SourceChecks)
	behaviorProblems, behaviorEvidence := validateBehaviors(m)
	problems = append(problems, validateRefs(m)...)
	problems = append(problems, sourceProblems...)
	problems = append(problems, behaviorProblems...)
	return problems, sourceEvidence, behaviorEvidence
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
	if len(m.SearchOrder) == 0 || len(m.Rules) == 0 || len(m.SourceChecks) == 0 {
		problems = append(problems, problem{Message: "manifest needs search_order, rules, and source_checks"})
	}
	return problems
}
