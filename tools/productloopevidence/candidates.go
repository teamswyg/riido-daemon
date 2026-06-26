package main

func collectCandidates(meta metaComplexity, product productAcceptance, partial partialReduction) []closedLoopCandiate {
	var out []closedLoopCandiate
	if meta.Status == statusPartial {
		out = append(out, closedLoopCandiate{
			ID:     "meta-complexity-entrypoint-budget",
			Class:  "meta-complexity",
			Reason: meta.PartialReason,
			RequiredNextArtifacts: []string{
				"claim-bound entrypoint index",
				"new contributor route map",
			},
		})
	}
	for _, id := range product.MissingSignalIDs {
		out = append(out, closedLoopCandiate{
			ID:     "product-outcome-signal-" + id,
			Class:  "product-acceptance",
			Reason: "outcome signal is not linked to local acceptance scenarios",
			RequiredNextArtifacts: []string{
				"local acceptance scenario",
				"product outcome evidence field",
			},
		})
	}
	if partial.Status == statusPartial {
		out = append(out, closedLoopCandiate{
			ID:     "partial-reduction-candidate-aging",
			Class:  "partial-reduction",
			Reason: "partial candidates need age, promotion, and escalation evidence",
			RequiredNextArtifacts: []string{
				"candidate created_at field",
				"candidate promotion verifier",
				"stale partial escalation evidence",
			},
		})
	}
	return out
}
