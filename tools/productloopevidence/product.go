package main

func buildProductAcceptance(m manifest, local localAcceptanceSource, run productRunOutcomeSource) productAcceptance {
	scenarios := scenarioSet(local.Scenarios)
	out := productAcceptance{
		ScenarioCount:           len(local.Scenarios),
		OutcomeSignalCount:      len(m.OutcomeSignals),
		LocalQARunEvidenceState: run.State,
		CoverageStatus:          run.CoverageStatus,
		DeploymentGateStatus:    run.DeploymentGateStatus,
		Status:                  statusPassed,
	}
	for _, signal := range m.OutcomeSignals {
		measure := buildOutcomeMeasure(signal, scenarios, run)
		out.MeasurementCandidates = append(out.MeasurementCandidates, measure)
		if measure.Linked {
			out.LinkedSignalCount++
		} else {
			out.MissingSignalIDs = append(out.MissingSignalIDs, signal.ID)
		}
		if measure.OutcomeEvidenceLinked {
			out.OutcomeEvidenceLinkedCount++
		} else {
			out.MissingOutcomeEvidenceSignalIDs = append(out.MissingOutcomeEvidenceSignalIDs, signal.ID)
		}
	}
	if len(out.MissingSignalIDs) > 0 || len(out.MissingOutcomeEvidenceSignalIDs) > 0 {
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
