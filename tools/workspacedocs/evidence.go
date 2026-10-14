package main

func buildEvidence(m manifest, results []sourceCheckResult, problems []string) evidence {
	return evidence{
		SchemaVersion: "riido-workspace-docs-result.v1",
		ManifestID:    m.ID,
		GeneratedDoc:  m.GeneratedDoc,
		Workflow:      m.Workflow,
		SourceChecks:  results,
		Problems:      problems,
	}
}
