package main

func validate(repo string, m Manifest) ([]problem, []SourceCheckResult, []AbsentCheck) {
	var problems []problem
	problems = append(problems, validateHeader(m)...)
	problems = append(problems, validateRefs(m)...)
	sourceResults := validateSources(repo, m.SourceChecks, &problems)
	absentResults := validateAbsent(repo, m.AbsentSurfaces, &problems)
	return problems, sourceResults, absentResults
}

func validateHeader(m Manifest) []problem {
	var problems []problem
	if m.SchemaVersion != "riido-validation-evidence.v1" {
		problems = append(problems, problem{"schema_version must be riido-validation-evidence.v1"})
	}
	for _, value := range []string{m.ID, m.Title, m.GeneratedDoc, m.Workflow, m.EvidenceArtifact, m.Purpose} {
		if value == "" {
			problems = append(problems, problem{"id, title, generated_doc, workflow, evidence_artifact, and purpose are required"})
		}
	}
	if len(m.Facts) == 0 || len(m.Boundaries) == 0 || len(m.SourceChecks) == 0 {
		problems = append(problems, problem{"facts, boundaries, and source_checks are required"})
	}
	return problems
}
