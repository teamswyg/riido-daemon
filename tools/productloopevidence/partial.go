package main

import "time"

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
