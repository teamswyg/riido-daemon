package main

import "time"

func candidateAges(items []registryLoop, staleAfterDays int, now time.Time) ([]candidateAge, int, int) {
	var out []candidateAge
	unknown := 0
	staleCount := 0
	for _, item := range items {
		if item.Kind == "closed-loop" {
			continue
		}
		created, err := time.Parse("2006-01-02", item.CandidateCreatedAt)
		if err != nil {
			unknown++
			continue
		}
		ageDays := int(now.Sub(created).Hours() / 24)
		stale := ageDays >= staleAfterDays
		if stale {
			staleCount++
		}
		out = append(out, candidateAge{
			ID:              item.ID,
			CreatedAt:       item.CandidateCreatedAt,
			AgeDays:         ageDays,
			PromotionTarget: item.PromotionTarget,
			Stale:           stale,
		})
	}
	return out, unknown, staleCount
}
