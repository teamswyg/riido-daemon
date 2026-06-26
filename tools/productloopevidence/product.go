package main

func buildProductAcceptance(m manifest, local localAcceptanceSource) productAcceptance {
	scenarios := scenarioSet(local.Scenarios)
	out := productAcceptance{
		ScenarioCount:      len(local.Scenarios),
		OutcomeSignalCount: len(m.OutcomeSignals),
		Status:             statusPassed,
	}
	for _, signal := range m.OutcomeSignals {
		linked := allPresent(signal.ScenarioIDs, scenarios)
		out.MeasurementCandidates = append(out.MeasurementCandidates, outcomeMeasure{
			ID:          signal.ID,
			ScenarioIDs: signal.ScenarioIDs,
			Linked:      linked,
		})
		if linked {
			out.LinkedSignalCount++
		} else {
			out.MissingSignalIDs = append(out.MissingSignalIDs, signal.ID)
		}
	}
	if len(out.MissingSignalIDs) > 0 {
		out.Status = statusPartial
	}
	return out
}

func scenarioSet(items []coverageScenario) map[string]bool {
	out := make(map[string]bool, len(items))
	for _, item := range items {
		out[item.ID] = true
	}
	return out
}

func allPresent(ids []string, set map[string]bool) bool {
	if len(ids) == 0 {
		return false
	}
	for _, id := range ids {
		if !set[id] {
			return false
		}
	}
	return true
}
