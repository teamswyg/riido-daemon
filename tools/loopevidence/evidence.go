package main

func buildEvidence(m manifest, docPath string, problems []string) evidenceReport {
	return evidenceReport{
		SchemaVersion:           "riido-loop-evidence-result.v1",
		ID:                      m.ID,
		Status:                  statusFor(problems),
		GeneratedDoc:            docPath,
		LoopCount:               len(m.Loops),
		RegisteredLoopFileCount: len(m.LoopFiles),
		OpenGapCount:            len(m.OpenGaps),
		EvidenceItemCount:       evidenceItemCount(m.Loops),
		RequiredPhases:          m.RequiredPhases,
		PhaseCoverage:           phaseCoverageRows(m),
		ProblemCount:            len(problems),
		ProblemSummaries:        append([]string{}, problems...),
	}
}

func evidenceItemCount(loops []loop) int {
	total := 0
	for _, item := range loops {
		total += len(item.Evidence)
	}
	return total
}

func statusFor(problems []string) string {
	if len(problems) > 0 {
		return "failed"
	}
	return "verified"
}
