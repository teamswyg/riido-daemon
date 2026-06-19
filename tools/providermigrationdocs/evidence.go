package main

func buildEvidence(m manifest, problems []string) evidence {
	return evidence{
		SchemaVersion:    "riido-provider-public-migration-docs-result.v1",
		ID:               m.ID,
		Status:           statusFor(problems),
		GeneratedDocs:    generatedDocPaths(m),
		PageCount:        len(m.Pages) + 1,
		ProviderCount:    providerCount(m.Pages),
		ArtifactCount:    artifactCount(m.Pages),
		Assertions:       m.Assertions,
		ProblemSummaries: problems,
		EvidenceArtifact: m.EvidenceArtifact,
	}
}

func providerCount(pages []page) int {
	count := 0
	for _, page := range pages {
		if page.ProviderID != "" {
			count++
		}
	}
	return count
}

func artifactCount(pages []page) int {
	count := 0
	for _, page := range pages {
		count += len(page.Artifacts)
	}
	return count
}
