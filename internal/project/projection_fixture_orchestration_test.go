package project

import "github.com/teamswyg/riido-daemon/internal/mwsdbridge"

func sampleSnapshotOrchestration() mwsdbridge.OrchestrationSnapshot {
	return mwsdbridge.OrchestrationSnapshot{
		SchemaVersion:          mwsdbridge.OrchestrationSchemaVersion,
		Root:                   "/workspace",
		DomainPath:             "/workspace/domains/macmini-workspace.lisp",
		HarnessRunPath:         "/workspace/harness/runs.jsonl",
		DomainSchemaVersion:    mwsdbridge.DomainSchemaVersion,
		HarnessSchemaVersion:   mwsdbridge.HarnessSchemaVersion,
		Mode:                   "orchestration-over-choreography",
		DecisionGate:           "human-approval-required",
		DecisionBy:             []string{"human"},
		DecisionLLMs:           []string{"codex"},
		RecommendedProvider:    "codex",
		RecommendedDecisionLLM: "codex",
		ProviderCandidates:     sampleSnapshotProviderCandidates(),
		NextAction: mwsdbridge.OrchestrationNextAction{
			Direction:             "top-down",
			CommandSurface:        "mwsd harness + riido task queue + mws-viewer cockpit",
			Reason:                "lift the latest bottom-up evidence into the next SSOT plan",
			RequiresHumanApproval: true,
		},
		TopDownCount:  1,
		BottomUpCount: 1,
		LastDirection: "bottom-up",
		Balanced:      true,
		DirectionBias: false,
		RecentRuns: []mwsdbridge.OrchestrationRun{
			{ID: "run-1", Direction: "top-down", Source: "docs/ROADMAP.md", Provider: "codex", Command: "plan", Result: "passed"},
		},
	}
}

func sampleSnapshotProviderCandidates() []mwsdbridge.ProviderCandidate {
	return []mwsdbridge.ProviderCandidate{
		{ID: "codex", SourceWorkflow: "provider-selection", Available: true, ApprovalRequired: true},
		{ID: "claude-code", SourceWorkflow: "provider-selection", Available: true, ApprovalRequired: true},
		{ID: "cursor", SourceWorkflow: "provider-selection", Available: true, ApprovalRequired: true},
	}
}
