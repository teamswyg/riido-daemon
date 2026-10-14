package main

import "time"

func buildEvidence(m manifest, commands []commandEvidence) evidenceFile {
	status := "verified"
	if len(commands) > 0 {
		status = "passed"
	}
	if anyFailed(commands) {
		status = "failed"
	}
	return evidenceFile{
		SchemaVersion: "riido-repo-verification-result.v1",
		ID:            m.ID,
		ObservedAt:    time.Now().UTC().Format(time.RFC3339),
		Status:        status,
		Commands:      commands,
		Assertions:    m.Assertions,
	}
}
