package main

func buildOutcomeMeasure(
	signal outcomeSignal,
	scenarios map[string]bool,
	run productRunOutcomeSource,
) outcomeMeasure {
	out := outcomeMeasure{ID: signal.ID, ScenarioIDs: signal.ScenarioIDs}
	for _, id := range signal.ScenarioIDs {
		present := scenarios[id]
		linked := run.State == localQARunFresh && acceptedOutcomeStatus(run.ScenarioStatus[id])
		out.ScenarioEvidence = append(out.ScenarioEvidence, outcomeScenarioEvidence{
			ID:                     id,
			LocalAcceptancePresent: present,
			RunStatus:              run.ScenarioStatus[id],
			OutcomeEvidenceLinked:  linked,
		})
		if !present {
			out.MissingScenarioIDs = append(out.MissingScenarioIDs, id)
		}
		if !linked {
			out.MissingOutcomeEvidenceScenarioIDs = append(out.MissingOutcomeEvidenceScenarioIDs, id)
		}
	}
	out.Linked = len(signal.ScenarioIDs) > 0 && len(out.MissingScenarioIDs) == 0
	out.OutcomeEvidenceLinked = len(signal.ScenarioIDs) > 0 &&
		len(out.MissingOutcomeEvidenceScenarioIDs) == 0
	return out
}
