package main

func collectProductOutcomeEvidenceCandidates(product productAcceptance) []closedLoopCandiate {
	var out []closedLoopCandiate
	for _, id := range product.MissingOutcomeEvidenceSignalIDs {
		out = append(out, closedLoopCandiate{
			ID:     "product-outcome-evidence-" + id,
			Class:  "product-acceptance",
			Reason: "outcome signal is declared but latest local QA run did not observe all bound scenarios",
			RequiredNextArtifacts: []string{
				"fresh local QA run evidence",
				"scenario coverage row with passed or observed status",
			},
			Graph: productOutcomeEvidenceGraph(id),
		})
	}
	return out
}
