package main

func candidateEvidenceGraphFor(candidate closedLoopCandidate) candidateEvidenceGraph {
	return candidateEvidenceGraph{
		Observation: candidate.Summary,
		Hypothesis:  candidateHypothesis(candidate),
		Change:      "Add or update the smallest closed-loop verifier that would catch this finding.",
		Verifier:    candidateVerifier(candidate),
		Evidence:    firstNonEmpty(candidate.Evidence, "local-qa-run.json"),
		Decision:    candidateDecision(candidate),
		NextLoop:    "closed-loop." + stableID(candidate.ID),
	}
}

func candidateHypothesis(candidate closedLoopCandidate) string {
	return "If " + candidate.Trigger + " remains only as harness output, the failure can recur."
}

func candidateVerifier(candidate closedLoopCandidate) string {
	return "Verify " + candidate.ID + " through a runner-backed test, gate, or explicit waiver."
}

func candidateDecision(candidate closedLoopCandidate) string {
	if candidate.Promoted {
		return "promoted_to_closed_loop"
	}
	if candidate.Stale {
		return "escalate_stale_partial"
	}
	return "candidate_for_promotion"
}
