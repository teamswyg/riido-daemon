package main

import "github.com/teamswyg/riido-daemon/internal/mwsdbridge"

func cliMwsdOrchestration() mwsdbridge.OrchestrationSnapshot {
	return mwsdbridge.OrchestrationSnapshot{
		SchemaVersion:          mwsdbridge.OrchestrationSchemaVersion,
		Root:                   cliMwsdRoot(),
		DomainSchemaVersion:    mwsdbridge.DomainSchemaVersion,
		HarnessSchemaVersion:   mwsdbridge.HarnessSchemaVersion,
		Mode:                   "human-gated-provider-selection",
		DecisionGate:           "human-approval-required",
		DecisionBy:             []string{"codex"},
		DecisionLLMs:           []string{"codex"},
		ProviderCandidates:     []mwsdbridge.ProviderCandidate{{ID: "codex", SourceWorkflow: "provider-selection", Available: true, ApprovalRequired: true}},
		RecommendedProvider:    "codex",
		RecommendedDecisionLLM: "codex",
		NextAction: mwsdbridge.OrchestrationNextAction{
			Direction:             "bottom-up",
			CommandSurface:        "riido task queue",
			Reason:                "continue migration",
			RequiresHumanApproval: true,
		},
		TopDownCount:  1,
		BottomUpCount: 0,
		LastDirection: "top-down",
		Balanced:      true,
	}
}
