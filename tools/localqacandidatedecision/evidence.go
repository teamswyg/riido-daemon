package main

import "time"

type evidence struct {
	SchemaVersion    string       `json:"schema_version"`
	ID               string       `json:"id"`
	GeneratedAt      string       `json:"generated_at"`
	EvidenceArtifact string       `json:"evidence_artifact"`
	Workflow         string       `json:"workflow"`
	Result           verifyResult `json:"result"`
}

func newEvidence(m manifest, result verifyResult) evidence {
	return evidence{
		SchemaVersion:    evidenceSchema,
		ID:               m.ID,
		GeneratedAt:      time.Now().UTC().Format(time.RFC3339),
		EvidenceArtifact: m.EvidenceArtifact,
		Workflow:         m.Workflow,
		Result:           result,
	}
}
