package main

func buildEvidence(m manifest, checks []sourceCheckResult, problems []string) evidence {
	return evidence{
		SchemaVersion:    "riido-distribution-host-docs-result.v1",
		ID:               m.ID,
		Status:           statusFor(problems),
		GeneratedDocs:    generatedDocPaths(m),
		SourceChecks:     checks,
		AssertionCount:   len(m.Assertions),
		PageCount:        len(generatedDocPaths(m)),
		ProblemSummaries: problems,
		EvidenceArtifact: m.EvidenceArtifact,
	}
}
