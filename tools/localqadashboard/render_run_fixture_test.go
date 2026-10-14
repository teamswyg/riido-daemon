package main

func renderTestRunEvidence() localRunEvidence {
	return localRunEvidence{
		ObservedAt:     "2026-06-22T01:00:00Z",
		ExpiresAt:      "2999-06-23T01:00:00Z",
		Status:         "passed",
		CoverageStatus: "partial",
		OpenRepairs:    []repairEvidence{renderTestRepair()},
		Candidates:     renderTestCandidates(),
	}
}

func renderTestRepair() repairEvidence {
	return repairEvidence{
		ProviderID:       "cursor",
		Class:            "provider_auth_required",
		Owner:            "human",
		Mode:             "manual",
		Summary:          "login required",
		SuggestedCommand: "cursor-agent login",
	}
}

func renderTestCandidates() []closedLoopCandidate {
	return []closedLoopCandidate{
		renderTestPendingCandidate(),
		renderTestPromotedCandidate(),
	}
}
