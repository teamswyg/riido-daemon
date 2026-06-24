package main

func closedLoopCandidate(id, source, class, reason, next string) evidenceGapCandidate {
	candidate := evidenceGapCandidate{
		ID:             id,
		SourceScenario: source,
		Class:          class,
		Reason:         reason,
		NextEvidence:   next,
	}
	candidate.Graph = candidateEvidenceGraph(candidate)
	return candidate
}

func candidateEvidenceGraph(candidate evidenceGapCandidate) evidenceGapCandidateGraph {
	source := candidate.SourceScenario
	if source == "" {
		source = "static-local-qa-gap"
	}
	return evidenceGapCandidateGraph{
		Observation: candidate.Reason,
		Hypothesis:  "A focused verifier or repair loop can remove this " + candidate.Class + " from local QA.",
		Change:      candidate.NextEvidence,
		Verifier:    "local.qa.evidence_gap_candidates",
		Evidence:    candidate.ID,
		Decision:    "Keep the candidate open until a claim-bound verifier or repair loop replaces the gap.",
		NextLoop:    "promote-" + source + "-" + candidate.ID,
	}
}
