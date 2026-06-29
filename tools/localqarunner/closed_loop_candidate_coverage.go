package main

func appendCoverageCandidates(out []closedLoopCandidate, seen map[string]bool,
	evidence runEvidence,
) []closedLoopCandidate {
	if evidence.Coverage == nil {
		return out
	}
	for _, row := range evidence.Coverage.Rows {
		if row.Status == statusPassed {
			continue
		}
		out = appendCandidate(out, seen, closedLoopCandidate{
			ID:         "coverage." + stableID(row.ID),
			Source:     "coverage",
			Trigger:    "coverage_not_passed",
			Summary:    coverageCandidateSummary(row),
			Evidence:   firstNonEmpty(row.Evidence, evidence.Artifacts.CoverageEvidence),
			NextAction: "Attach fresh evidence or promote this coverage gap into a closed-loop verifier.",
		})
	}
	return out
}

func coverageCandidateSummary(row runCoverageRow) string {
	if row.Repair != nil && row.Repair.Summary != "" {
		return row.Repair.Summary
	}
	if row.Detail != "" {
		return row.Detail
	}
	return row.Title + " is not fully verified"
}
