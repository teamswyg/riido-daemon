package main

func buildReport(reg registry, changed *changedSummary, problems []string) report {
	out := report{
		SchemaVersion:      reportVersion,
		ID:                 reg.ID,
		Status:             statusPassed,
		GeneratedDoc:       reg.GeneratedDoc,
		Workflow:           reg.Workflow,
		EvidenceArtifact:   reg.EvidenceArtifact,
		LoopCount:          len(reg.Loops),
		BusinessClaimCount: len(reg.BusinessClaims),
		ProblemCount:       len(problems),
		Problems:           problems,
		Loops:              loopSummaries(reg.Loops),
		BusinessClaims:     claimSummaries(reg.BusinessClaims),
	}
	out.ChangedFileCheck = changed
	if len(problems) > 0 {
		out.Status = statusFailed
	}
	return out
}

func loopSummaries(items []loopEntry) []loopSummary {
	out := make([]loopSummary, 0, len(items))
	for _, item := range items {
		out = append(out, loopSummary{
			ID:           item.ID,
			Kind:         item.Kind,
			ExpiresAfter: item.ExpiresAfter,
			Evidence:     item.Evidence,
		})
	}
	return out
}

func claimSummaries(items []businessClaim) []claimSummary {
	out := make([]claimSummary, 0, len(items))
	for _, item := range items {
		out = append(out, claimSummary{
			ID:            item.ID,
			FileCount:     len(item.Files),
			DocCount:      len(item.Docs),
			EvidenceCount: len(item.Evidence),
			VerifierCount: len(item.Verifiers),
			BoundFiles:    item.Files,
		})
	}
	return out
}
