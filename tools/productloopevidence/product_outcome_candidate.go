package main

import "strings"

func collectProductOutcomeEvidenceCandidates(product productAcceptance) []closedLoopCandiate {
	var out []closedLoopCandiate
	for _, id := range product.MissingOutcomeEvidenceSignalIDs {
		missing := missingOutcomeEvidenceScenarioIDs(product, id)
		out = append(out, closedLoopCandiate{
			ID:     "product-outcome-evidence-" + id,
			Class:  "product-acceptance",
			Reason: "outcome signal is declared but latest local QA run did not observe all bound scenarios",
			RequiredNextArtifacts: []string{
				"fresh local QA run evidence",
				"scenario coverage row with passed or observed status",
				"missing scenario ids: " + strings.Join(missing, ", "),
			},
			Graph: productOutcomeEvidenceGraph(id),
		})
	}
	return out
}

func missingOutcomeEvidenceScenarioIDs(product productAcceptance, id string) []string {
	for _, measure := range product.MeasurementCandidates {
		if measure.ID == id {
			return measure.MissingOutcomeEvidenceScenarioIDs
		}
	}
	return nil
}
