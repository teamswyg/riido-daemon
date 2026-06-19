package main

func buildEvidence(
	manifest Manifest,
	problems []problem,
	manifestChecks []ManifestCheck,
	sourceChecks []SourceResult,
) Evidence {
	status := "verified"
	if len(problems) > 0 {
		status = "failed"
	}
	return Evidence{
		SchemaVersion:    "riido-approval-wait-timeout-result.v1",
		ID:               manifest.ID,
		Status:           status,
		ManifestChecks:   manifestChecks,
		SourceChecks:     sourceChecks,
		Assertions:       manifest.Assertions,
		ProblemSummaries: problemMessages(problems),
		EvidenceArtifact: manifest.EvidenceArtifact,
	}
}
