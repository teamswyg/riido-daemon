package main

func applyCandidatePromotions(candidates []closedLoopCandidate,
	promotions []closedLoopPromotion,
) []closedLoopCandidate {
	byID := promotionByCandidate(promotions)
	for i := range candidates {
		promotion, ok := byID[candidates[i].ID]
		if !ok {
			continue
		}
		candidates[i].Promoted = true
		candidates[i].Stale = false
		candidates[i].Status = "promoted"
		candidates[i].Graph = promotedCandidateGraph(candidates[i], promotion)
	}
	return candidates
}

func promotedCandidateGraph(candidate closedLoopCandidate,
	promotion closedLoopPromotion,
) candidateEvidenceGraph {
	graph := candidateEvidenceGraphFor(candidate)
	graph.Verifier = firstNonEmpty(promotion.Verifier, graph.Verifier)
	graph.Evidence = firstNonEmpty(promotion.Evidence, graph.Evidence)
	graph.Decision = firstNonEmpty(promotion.Decision, "promoted_to_closed_loop")
	graph.NextLoop = firstNonEmpty(promotion.LoopSource, "closed-loop."+stableID(promotion.LoopID))
	return graph
}
