package main

func buildEvidence(m manifest, checks []sourceCheckResult, problems []string) evidence {
	return evidence{
		SchemaVersion:    "riido-provider-runtime-boundary-docs-result.v1",
		ID:               m.ID,
		Status:           statusFor(problems),
		GeneratedDocs:    generatedDocPaths(m),
		AssertionCount:   len(m.Assertions),
		DetailCount:      len(m.Details),
		SourceChecks:     checks,
		ProblemSummaries: problems,
		EvidenceArtifact: m.EvidenceArtifact,
	}
}
