package main

func buildEvidence(
	manifest Manifest,
	problems []problem,
	interfaces []InterfaceCheck,
	sources []SourceResult,
) Evidence {
	status := "verified"
	if len(problems) > 0 {
		status = "failed"
	}
	return Evidence{
		SchemaVersion:    "riido-process-lifecycle-result.v1",
		ID:               manifest.ID,
		Status:           status,
		InterfaceChecks:  interfaces,
		SourceChecks:     sources,
		Assertions:       manifest.Assertions,
		ProblemSummaries: problemMessages(problems),
		EvidenceArtifact: manifest.EvidenceArtifact,
	}
}
