package main

const providerRepairCandidateMode = "candidate_auto"

func appendProviderRepairCandidates(evidence *runEvidence, providers []providerRunProvider) {
	for _, provider := range providers {
		repair := providerRepair(provider)
		if repair == nil || repair.Mode != providerRepairCandidateMode {
			continue
		}
		evidence.ClosedLoops = append(evidence.ClosedLoops, providerRepairCandidate(provider.ID, *repair))
	}
}

func providerRepairCandidate(providerID string, repair runRepair) runLoopCandidate {
	id := "repair-provider." + providerID + "." + repair.Class
	reason := "Provider " + providerID + " repair required: " + repair.Summary
	next := repair.SuggestedCommand
	if next == "" {
		next = "capture repair evidence and rerun provider integration"
	}
	candidate := runLoopCandidate{
		ID:                    id,
		SourceScenario:        "provider.real_cli_observation",
		Class:                 repair.Class,
		Reason:                reason,
		NextEvidence:          next,
		RequiredNextArtifacts: providerRepairCandidateArtifacts(),
	}
	candidate.Graph = providerRepairCandidateGraph(candidate)
	return candidate
}

func providerRepairCandidateArtifacts() []string {
	return []string{
		"claim_binding",
		"verifier",
		"ci_gate",
		"redacted_evidence",
		"decision_record",
		"evidence_graph_edge",
	}
}
