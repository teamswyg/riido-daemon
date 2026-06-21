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
		ClosedCount:   len(m.ClosedLoops),
		Problems:      []string{},
		Checks:        []checkSummary{},
		ClosedLoops:   []closedSummary{},
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

func summarizeClosedLoops(m manifest, checks []checkSummary) ([]closedSummary, []string) {
	failed := failedEvidence(checks)
	var out []closedSummary
	var problems []string
	for _, item := range m.ClosedLoops {
		status := statusVerified
		for _, id := range item.EvidenceIDs {
			if failed[id] {
				status = statusFailed
				problems = append(problems, item.ID+" is open because "+id+" failed")
			}
		}
		out = append(out, closedSummary{ID: item.ID, Kind: item.Kind, Status: status})
	}
	return out, problems
}

func failedEvidence(checks []checkSummary) map[string]bool {
	failed := map[string]bool{}
	for _, check := range checks {
		if check.Status != statusVerified {
			failed[check.EvidenceID] = true
		}
	}
	return failed
}
