package main

import (
	"os"
	"time"
)

func buildPartialReduction(root string, m manifest, reg registrySource, qa qaSystemSource) partialReduction {
	return buildPartialReductionAt(root, m, reg, qa, time.Now().UTC())
}

func buildPartialReductionAt(root string, m manifest, reg registrySource, qa qaSystemSource, now time.Time) partialReduction {
	candidateIDs := candidateLoopIDs(reg.Loops)
	inferred := inferenceRequiredIDs(qa.ExecutionInventory)
	localEvidence := localQARunPresent(root, m.LocalQARunEvidence)
	ages, unknown, stale := candidateAges(reg.Loops, m.Thresholds.StalePartialAfterDays, now)
	out := partialReduction{
		InferenceRequiredIDs:      inferred,
		ClosedLoopCandidateIDs:    candidateIDs,
		CandidateAges:             ages,
		CandidateCount:            len(candidateIDs),
		PromotedCount:             promotedCount(reg.Loops),
		CandidateAgeUnknownCount:  unknown,
		StaleCandidateCount:       stale,
		StalePartialAfterDays:     m.Thresholds.StalePartialAfterDays,
		LocalQARunEvidencePresent: localEvidence,
		Status:                    statusPassed,
	}
	if len(inferred) > 0 || len(candidateIDs) > 0 || !localEvidence {
		out.Status = statusPartial
	}
	return out
}

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

func inferenceRequiredIDs(items []qaExecution) []string {
	var out []string
	for _, item := range items {
		if item.Mode != "" && item.Mode != "system" {
			out = append(out, item.ID)
		}
	}
	return out
}

func candidateLoopIDs(items []registryLoop) []string {
	var out []string
	for _, item := range items {
		if item.Kind != "closed-loop" {
			out = append(out, item.ID)
		}
	}
	return out
}

func promotedCount(items []registryLoop) int {
	count := 0
	for _, item := range items {
		if item.Kind == "closed-loop" {
			count++
		}
	}
	return count
}

func localQARunPresent(root, rel string) bool {
	if rel == "" {
		return false
	}
	_, err := os.Stat(repoPath(root, rel))
	return err == nil
}
