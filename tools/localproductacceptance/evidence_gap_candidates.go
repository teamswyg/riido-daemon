package main

func evidenceGapCandidates(
	items []scenario,
	manualPresent bool,
	capturePresent bool,
	captureUploadCovered bool,
) []evidenceGapCandidate {
	out := staticEvidenceGapCandidates(manualPresent, capturePresent, captureUploadCovered)
	out = append(out, scenarioEvidenceGapCandidates(items)...)
	return out
}

func legacyCandidateRows(candidates []evidenceGapCandidate) []map[string]string {
	out := make([]map[string]string, 0, len(candidates))
	for _, candidate := range candidates {
		out = append(out, map[string]string{
			"id":            candidate.ID,
			"reason":        candidate.Reason,
			"next_evidence": candidate.NextEvidence,
		})
	}
	return out
}
