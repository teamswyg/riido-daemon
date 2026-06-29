package main

func scenarioEvidenceGapCandidates(items []scenario) []evidenceGapCandidate {
	out := []evidenceGapCandidate{}
	for _, item := range items {
		out = append(out, inferredExecutionCandidates(item)...)
		if item.Status == statusSkipped {
			out = append(out, skippedScenarioCandidate(item))
		}
		if item.Status == statusPartial {
			out = append(out, partialScenarioCandidate(item))
		}
		if item.Status == statusFailed {
			out = append(out, failedScenarioCandidate(item))
		}
	}
	return out
}

func skippedScenarioCandidate(item scenario) evidenceGapCandidate {
	if item.Repair != nil {
		return repairedScenarioCandidate(item)
	}
	return scenarioGap(item, "skipped_probe", "Scenario was skipped by the harness.")
}

func partialScenarioCandidate(item scenario) evidenceGapCandidate {
	if item.Repair != nil {
		return repairedScenarioCandidate(item)
	}
	return scenarioGap(item, "partial_evidence", "Scenario produced partial evidence.")
}

func failedScenarioCandidate(item scenario) evidenceGapCandidate {
	if item.Repair != nil {
		return repairedScenarioCandidate(item)
	}
	return scenarioGap(item, "failed_probe", "Scenario failed during local QA.")
}

func repairedScenarioCandidate(item scenario) evidenceGapCandidate {
	next := item.Repair.SuggestedCommand
	if next == "" {
		next = "Execute the scenario repair and rerun local QA."
	}
	return closedLoopCandidate("repair-"+item.ID, item.ID, item.Repair.Class, item.Repair.Summary, next)
}

func scenarioGap(item scenario, class, reason string) evidenceGapCandidate {
	return closedLoopCandidate(
		"close-"+item.ID,
		item.ID,
		class,
		reason,
		"Promote the observed failure into a focused verifier or repair classifier.",
	)
}
