package main

func buildEvidence(manifest Manifest, problems []problem, fields, builder []FieldCheck) Evidence {
	status := "verified"
	if len(problems) > 0 {
		status = "failed"
	}
	return Evidence{
		SchemaVersion:    "riido-event-authority-result.v1",
		ID:               manifest.ID,
		Status:           status,
		FieldChecks:      fields,
		BuilderChecks:    builder,
		ProblemSummaries: problemMessages(problems),
		EvidenceArtifact: manifest.EvidenceArtifact,
	}
}
