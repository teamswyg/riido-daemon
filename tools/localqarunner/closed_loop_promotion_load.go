package main

import (
	"encoding/json"
	"os"
)

func loadClosedLoopPromotions(path string) []closedLoopPromotion {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var registry closedLoopPromotionRegistry
	if json.Unmarshal(data, &registry) != nil {
		return nil
	}
	if registry.SchemaVersion != "riido-local-qa-closed-loop-promotions.v1" {
		return nil
	}
	return registry.Promotions
}

func promotionByCandidate(promotions []closedLoopPromotion) map[string]closedLoopPromotion {
	out := map[string]closedLoopPromotion{}
	for _, promotion := range promotions {
		if promotion.CandidateID != "" {
			out[promotion.CandidateID] = promotion
		}
	}
	return out
}
