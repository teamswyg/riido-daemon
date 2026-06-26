package main

import "os"

func buildPartialReduction(root string, m manifest, reg registrySource, qa qaSystemSource) partialReduction {
	candidateIDs := candidateLoopIDs(reg.Loops)
	inferred := inferenceRequiredIDs(qa.ExecutionInventory)
	localEvidence := localQARunPresent(root, m.LocalQARunEvidence)
	out := partialReduction{
		InferenceRequiredIDs:      inferred,
		ClosedLoopCandidateIDs:    candidateIDs,
		CandidateCount:            len(candidateIDs),
		PromotedCount:             promotedCount(reg.Loops),
		CandidateAgeUnknownCount:  len(candidateIDs),
		StalePartialAfterDays:     m.Thresholds.StalePartialAfterDays,
		LocalQARunEvidencePresent: localEvidence,
		Status:                    statusPassed,
	}
	if len(inferred) > 0 || len(candidateIDs) > 0 || !localEvidence {
		out.Status = statusPartial
	}
	return out
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
