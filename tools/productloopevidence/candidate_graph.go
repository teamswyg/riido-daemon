package main

func metaCandidateGraph() candidateGraph {
	return candidateGraph{
		Observation: "tool, workflow, and verifier entrypoints can exceed a human-readable surface",
		Hypothesis:  "every visible entrypoint should belong to a route map before meta complexity passes",
		Change:      "entrypoint route map coverage is part of product-loop evidence",
		Verifier:    "TestBuildMetaComplexityTreatsRoutedEntrypointsAsManaged",
		Evidence:    "product-loop-evidence.meta_complexity.route_coverage",
		Decision:    "unrouted entrypoints become product-loop partial debt",
		NextLoop:    "entrypoint-route-map",
	}
}

func productSignalGraph(id string) candidateGraph {
	return candidateGraph{
		Observation: "product outcome signal is declared without local acceptance scenario evidence",
		Hypothesis:  "product success cannot be proven unless the signal is bound to runnable QA",
		Change:      "productloopevidence checks outcome signal scenario bindings",
		Verifier:    "TestBuildProductAcceptanceFindsMissingSignals",
		Evidence:    "product-loop-evidence.product_acceptance." + id,
		Decision:    "missing outcome binding remains product acceptance debt",
		NextLoop:    "local-product-acceptance",
	}
}

func productOutcomeEvidenceGraph(id string) candidateGraph {
	return candidateGraph{
		Observation: "product outcome signal is bound to scenarios but the latest run evidence did not observe the outcome",
		Hypothesis:  "a product loop is not closed until local QA run rows prove the scenario actually executed",
		Change:      "productloopevidence reads local QA run coverage rows for outcome signal linkage",
		Verifier:    "TestBuildProductAcceptanceRequiresRunOutcomeEvidence",
		Evidence:    "product-loop-evidence.product_acceptance." + id + ".outcome_evidence_linked",
		Decision:    "missing outcome run evidence remains product acceptance debt",
		NextLoop:    "local-product-acceptance",
	}
}

func candidateAgingGraph() candidateGraph {
	return candidateGraph{
		Observation: "partial candidates can linger without promotion or escalation",
		Hypothesis:  "candidate age should become measurable stale partial evidence",
		Change:      "productloopevidence computes candidate age and stale count",
		Verifier:    "TestBuildPartialReductionComputesCandidateAge",
		Evidence:    "product-loop-evidence.partial_reduction.candidate_ages",
		Decision:    "stale partials must become escalation evidence",
		NextLoop:    "local-qa-candidate-decision",
	}
}

func localQARunGraph() candidateGraph {
	return candidateGraph{
		Observation: "candidate promotion can be detached from a concrete local QA run",
		Hypothesis:  "product-loop evidence should show whether latest run evidence exists",
		Change:      "productloopevidence checks local_qa_run_evidence presence",
		Verifier:    "TestCollectCandidatesPromotesPartialEvidence",
		Evidence:    "product-loop-evidence.partial_reduction.local_qa_run_evidence_present",
		Decision:    "absent run evidence keeps product-loop partial",
		NextLoop:    "local-qa-evidence-expiry-gate",
	}
}
