package main

func buildEvidence(
	m Manifest,
	problems []problem,
	sources []SourceCheckResult,
	absent []AbsentCheck,
) Evidence {
	status := "verified"
	if len(problems) > 0 {
		status = "failed"
	}
	return Evidence{
		SchemaVersion: "riido-validation-evidence-result.v1",
		ID:            m.ID,
		Status:        status,
		SourceChecks:  sources,
		AbsentChecks:  absent,
		Problems:      messages(problems),
	}
}
