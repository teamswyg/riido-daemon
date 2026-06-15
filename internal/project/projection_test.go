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
	if projection.SchemaVersion != "riido-workspace-projection.v1" {
		t.Fatalf("unexpected schema version: %s", projection.SchemaVersion)
	}
	if projection.Domain != "macmini-workspace" {
		t.Fatalf("unexpected domain: %s", projection.Domain)
	}
	if projection.DocumentCount != 23 {
		t.Fatalf("unexpected document count: %d", projection.DocumentCount)
	}
	if projection.HarnessNextDirection != "top-down" {
		t.Fatalf("unexpected harness next direction: %s", projection.HarnessNextDirection)
	}
	if projection.OrchestrationSchema != mwsdbridge.OrchestrationSchemaVersion {
		t.Fatalf("unexpected orchestration schema: %s", projection.OrchestrationSchema)
	}
	if projection.DecisionGate != "human-approval-required" {
		t.Fatalf("unexpected decision gate: %s", projection.DecisionGate)
	}
	if projection.RecommendedProvider != "codex" {
		t.Fatalf("unexpected recommended provider: %s", projection.RecommendedProvider)
	}
	if projection.RecommendedDecisionLLM != "codex" {
		t.Fatalf("unexpected recommended decision LLM: %s", projection.RecommendedDecisionLLM)
	}
	if len(projection.ProviderCandidates) != 3 {
		t.Fatalf("unexpected provider candidate count: %d", len(projection.ProviderCandidates))
	}
	if !projection.HarnessBalanced {
		t.Fatalf("expected harness to be balanced")
	}
	if len(projection.Projects) != 3 {
		t.Fatalf("unexpected project count: %d", len(projection.Projects))
	}
	if !projection.Ready() {
		t.Fatalf("projection should be ready, diagnostics=%v", projection.Diagnostics)
	}
	if projection.Projects[0].ID != "gui_engine" {
		t.Fatalf("projects should be sorted by id: %#v", projection.Projects)
	}
	if len(projection.DocumentTaskLinks) != 2 {
		t.Fatalf("unexpected document task link count: %d", len(projection.DocumentTaskLinks))
	}
	if projection.DocumentTaskLinks[0].TaskID != "task:mws.goal" {
		t.Fatalf("unexpected first task id: %s", projection.DocumentTaskLinks[0].TaskID)
	}
	if projection.DocumentTaskLinks[0].ProjectID != "macmini-workspace" {
		t.Fatalf("unexpected project id: %s", projection.DocumentTaskLinks[0].ProjectID)
	}
	if projection.DocumentTaskLinks[0].RecommendedProvider != "codex" {
		t.Fatalf("unexpected task recommended provider: %s", projection.DocumentTaskLinks[0].RecommendedProvider)
	}
	if projection.DocumentTaskLinks[0].RecommendedDecisionLLM != "codex" {
		t.Fatalf("unexpected task decision LLM: %s", projection.DocumentTaskLinks[0].RecommendedDecisionLLM)
	}
	if !projection.DocumentTaskLinks[0].RequiresHumanApproval {
		t.Fatalf("task link should require human approval")
	}
	for _, project := range projection.Projects {
		if project.Health != RepositoryReady {
			t.Fatalf("project %s should be ready, got %s", project.ID, project.Health)
		}
	}
}

func TestFromMwsdSnapshotReportsRepositoryHealth(t *testing.T) {
	snapshot := sampleSnapshot()
	snapshot.Projects.Repositories[1].RemoteMatches = false

	projection, err := FromMwsdSnapshot(snapshot)
	if err != nil {
		t.Fatalf("FromMwsdSnapshot returned error: %v", err)
	}
	if projection.Ready() {
		t.Fatal("projection should not be ready when a repo remote mismatches")
	}
	var found bool
	for _, project := range projection.Projects {
		if project.ID == "gui_engine" {
			found = true
			if project.Health != RepositoryRemoteMismatch {
				t.Fatalf("unexpected gui_engine health: %s", project.Health)
			}
		}
	}
	if !found {
		t.Fatal("gui_engine project missing")
	}
}
