package project

import (
	"sort"

	"github.com/teamswyg/riido-daemon/internal/mwsdbridge"
)

func providerCandidates(candidates []mwsdbridge.ProviderCandidate) []ProviderCandidate {
	out := make([]ProviderCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		out = append(out, ProviderCandidate{
			ID:               candidate.ID,
			SourceWorkflow:   candidate.SourceWorkflow,
			Available:        candidate.Available,
			ApprovalRequired: candidate.ApprovalRequired,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

func nextAction(action mwsdbridge.OrchestrationNextAction) NextAction {
	return NextAction{
		Direction:             action.Direction,
		CommandSurface:        action.CommandSurface,
		Reason:                action.Reason,
		RequiresHumanApproval: action.RequiresHumanApproval,
	}
}
