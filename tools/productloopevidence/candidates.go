package main

func collectCandidates(
	meta metaComplexity,
	product productAcceptance,
	schedule qaScheduleEvidence,
	partial partialReduction,
) []closedLoopCandiate {
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
			Graph: metaCandidateGraph(),
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
			Graph: productSignalGraph(id),
		})
	}
	if schedule.Status == statusPartial {
		out = append(out, scheduleCandidate(schedule))
	}
	if partial.CandidateAgeUnknownCount > 0 || partial.StaleCandidateCount > 0 {
		out = append(out, closedLoopCandiate{
			ID:     "partial-reduction-candidate-aging",
			Class:  "partial-reduction",
			Reason: "partial candidates need age, promotion, and escalation evidence",
			RequiredNextArtifacts: []string{
				"candidate created_at field",
				"candidate promotion verifier",
				"stale partial escalation evidence",
			},
			Graph: candidateAgingGraph(),
		})
	}
	if partial.CandidateCount > 0 && partial.CandidateAgeUnknownCount == 0 && partial.StaleCandidateCount == 0 {
		out = append(out, openCandidateDebt(partial))
	}
	if partial.CandidateCount > 0 && !partial.LocalQARunEvidenceFresh {
		out = append(out, closedLoopCandiate{
			ID:     "partial-reduction-local-qa-run-evidence",
			Class:  "partial-reduction",
			Reason: "local QA run evidence is " + partial.LocalQARunEvidenceState + ", so candidate promotion cannot be tied to a fresh concrete run",
			RequiredNextArtifacts: []string{
				"latest local QA run evidence",
				"candidate decision evidence",
			},
			Graph: localQARunGraph(),
		})
	}
	return out
}
