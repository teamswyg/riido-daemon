package main

func validate(repo string, m Manifest) (
	[]problem,
	[]SourceEvidence,
	[]GateEvidence,
	[]AbsentEvidence,
) {
	problems := validateRequired(m)
	sourceProblems, sourceEvidence := validateSources(repo, m.SourceChecks)
	gateProblems, gateEvidence := validateGates(m.Gates)
	absentProblems, absentEvidence := validateAbsent(repo, m.AbsentScans)
	problems = append(problems, validateRefs(m)...)
	problems = append(problems, sourceProblems...)
	problems = append(problems, gateProblems...)
	problems = append(problems, absentProblems...)
	return problems, sourceEvidence, gateEvidence, absentEvidence
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
	if len(m.Inputs) == 0 || len(m.Gates) == 0 || len(m.SourceChecks) == 0 {
		problems = append(problems, problem{Message: "manifest needs inputs, gates, and source checks"})
	}
	return problems
}
