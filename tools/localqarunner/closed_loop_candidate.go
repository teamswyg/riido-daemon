package main

const candidateStaleAfterHours = 72

func applyClosedLoopCandidates(evidence runEvidence) runEvidence {
	candidates := closedLoopCandidates(evidence)
	evidence.Candidates = candidates
	evidence.CandidateSummary = closedLoopSummary{
		Total:           len(candidates),
		StaleAfterHours: candidateStaleAfterHours,
	}
	return evidence
}

func closedLoopCandidates(evidence runEvidence) []closedLoopCandidate {
	seen := map[string]bool{}
	out := make([]closedLoopCandidate, 0)
	out = appendStepCandidates(out, seen, evidence)
	out = appendRepairCandidates(out, seen, evidence)
	out = appendCoverageCandidates(out, seen, evidence)
	return out
}

func appendCandidate(out []closedLoopCandidate, seen map[string]bool,
	candidate closedLoopCandidate,
) []closedLoopCandidate {
	if seen[candidate.ID] {
		return out
	}
	seen[candidate.ID] = true
	candidate.Status = "candidate"
	candidate.StaleAfterHours = candidateStaleAfterHours
	return append(out, candidate)
}
