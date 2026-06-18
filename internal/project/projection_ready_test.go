package project

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/mwsdbridge"
)

func TestFromMwsdSnapshotProjectsReady(t *testing.T) {
	projection, err := FromMwsdSnapshot(sampleSnapshot())
	if err != nil {
		t.Fatalf("FromMwsdSnapshot returned error: %v", err)
	}
	assertProjectionShape(t, projection)
	assertProjectionRecommendations(t, projection)
	assertProjectionProjects(t, projection)
	assertProjectionTaskLinks(t, projection)
}

func assertProjectionShape(t *testing.T, projection WorkspaceProjection) {
	t.Helper()
	if projection.SchemaVersion != "riido-workspace-projection.v1" ||
		projection.Domain != "macmini-workspace" ||
		projection.DocumentCount != 23 ||
		projection.HarnessNextDirection != "top-down" ||
		projection.OrchestrationSchema != mwsdbridge.OrchestrationSchemaVersion {
		t.Fatalf("unexpected projection shape: %+v", projection)
	}
	if projection.DecisionGate != "human-approval-required" || !projection.HarnessBalanced {
		t.Fatalf("unexpected decision state: %+v", projection)
	}
}

func assertProjectionRecommendations(t *testing.T, projection WorkspaceProjection) {
	t.Helper()
	if projection.RecommendedProvider != "codex" || projection.RecommendedDecisionLLM != "codex" {
		t.Fatalf("unexpected recommendations: %+v", projection)
	}
	if len(projection.ProviderCandidates) != 3 {
		t.Fatalf("unexpected provider candidate count: %d", len(projection.ProviderCandidates))
	}
}
