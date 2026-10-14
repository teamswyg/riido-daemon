package project

import "github.com/teamswyg/riido-daemon/internal/mwsdbridge"

func baseWorkspaceProjection(snapshot mwsdbridge.Snapshot) WorkspaceProjection {
	return WorkspaceProjection{
		SchemaVersion:           "riido-workspace-projection.v1",
		Root:                    snapshot.Status.Root,
		Domain:                  snapshot.Domain.Domain,
		DomainPath:              snapshot.Domain.Path,
		DocumentCount:           snapshot.Graph.Stats.DocumentCount,
		GraphNodeCount:          snapshot.Graph.Stats.NodeCount,
		GraphEdgeCount:          snapshot.Graph.Stats.EdgeCount,
		HarnessRunCount:         snapshot.Harness.RunCount,
		HarnessNextDirection:    snapshot.Harness.NextDirection,
		HarnessRecentDirections: append([]string(nil), snapshot.Harness.RecentDirections...),
		OrchestrationSchema:     snapshot.Orchestration.SchemaVersion,
		OrchestrationMode:       snapshot.Orchestration.Mode,
		DecisionGate:            snapshot.Orchestration.DecisionGate,
		DecisionBy:              append([]string(nil), snapshot.Orchestration.DecisionBy...),
		DecisionLLMs:            append([]string(nil), snapshot.Orchestration.DecisionLLMs...),
		ProviderCandidates:      providerCandidates(snapshot.Orchestration.ProviderCandidates),
		RecommendedProvider:     snapshot.Orchestration.RecommendedProvider,
		RecommendedDecisionLLM:  snapshot.Orchestration.RecommendedDecisionLLM,
		NextAction:              nextAction(snapshot.Orchestration.NextAction),
		HarnessBalanced:         snapshot.Orchestration.Balanced,
		DirectionBias:           snapshot.Orchestration.DirectionBias,
		RecentProviderRuns:      providerRuns(snapshot.Orchestration.RecentRuns),
		SSOTConflictCount:       snapshot.Status.SSOTConflictCount,
	}
}
