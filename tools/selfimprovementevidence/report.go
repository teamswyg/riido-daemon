package main

func newReport(m manifest) report {
	return report{
		SchemaVersion: evidenceSchema,
		ID:            m.ID,
		Status:        statusVerified,
		GeneratedDoc:  m.GeneratedDoc,
		Workflow:      m.Workflow,
		Artifact:      m.EvidenceArtifact,
		LoopSource:    m.LoopSource,
		RequiredCount: len(m.Required),
		Problems:      []string{},
		Checks:        []checkSummary{},
	}
}

func countVerifiedEvidence(checks []checkSummary, required []requiredEvidence) int {
	failed := map[string]bool{}
	seen := map[string]bool{}
	for _, check := range checks {
		seen[check.EvidenceID] = true
		if check.Status != statusVerified {
			failed[check.EvidenceID] = true
		}
	}
	count := 0
	for _, item := range required {
		if seen[item.ID] && !failed[item.ID] {
			count++
		}
	}
	return count
}
