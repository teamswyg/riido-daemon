package main

func openCandidateDebt(partial partialReduction) closedLoopCandiate {
	return closedLoopCandiate{
		ID:     "partial-reduction-open-candidate-debt",
		Class:  "partial-reduction",
		Reason: "open closed-loop candidate ids remain in the loop registry",
		RequiredNextArtifacts: []string{
			"candidate decision evidence",
			"promotion verifier",
			"closed-loop registry update",
		},
		Graph: openCandidateDebtGraph(partial),
	}
}

func openCandidateDebtGraph(partial partialReduction) candidateGraph {
	return candidateGraph{
		Observation: "loop registry still contains non-closed-loop candidate ids",
		Hypothesis:  "fresh QA evidence should still expose remaining candidate debt",
		Change:      "productloopevidence emits open candidate debt when candidate_count is non-zero",
		Verifier:    "TestCollectCandidatesShowsOpenCandidateDebt",
		Evidence:    "product-loop-evidence.partial_reduction.closed_loop_candidate_ids",
		Decision:    "candidate_count=" + itoa(partial.CandidateCount) + " keeps product-loop partial",
		NextLoop:    "local-qa-candidate-decision",
	}
}
