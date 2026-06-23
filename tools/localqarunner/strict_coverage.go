package main

func applyStrictCoverage(evidence runEvidence) runEvidence {
	if evidence.StrictCoverage && evidence.CoverageStatus != statusPassed {
		evidence.Status = statusFailed
	}
	return evidence
}
