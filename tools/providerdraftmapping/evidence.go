package main

func buildEvidence(manifest Manifest, problems []problem, mappings []MappingCheck, coverage []CoverageCheck) Evidence {
	status := "verified"
	if len(problems) > 0 {
		status = "failed"
	}
	return Evidence{
		SchemaVersion:    "riido-provider-draft-mapping-result.v1",
		ID:               manifest.ID,
		Status:           status,
		MappingChecks:    mappings,
		CoverageChecks:   coverage,
		Assertions:       manifest.Assertions,
		ProblemSummaries: problemMessages(problems),
		EvidenceArtifact: manifest.EvidenceArtifact,
	}
}
