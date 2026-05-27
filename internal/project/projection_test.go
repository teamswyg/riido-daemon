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

func sampleSnapshot() mwsdbridge.Snapshot {
	return mwsdbridge.Snapshot{
		Status: mwsdbridge.Status{
			Root:                       "/workspace",
			GraphSchemaVersion:         mwsdbridge.GraphSchemaVersion,
			DomainSchemaVersion:        mwsdbridge.DomainSchemaVersion,
			HarnessSchemaVersion:       mwsdbridge.HarnessSchemaVersion,
			OrchestrationSchemaVersion: mwsdbridge.OrchestrationSchemaVersion,
			DocumentCount:              23,
			RepositoryCount:            3,
			DomainName:                 "macmini-workspace",
			HarnessRunCount:            2,
			HarnessNextDirection:       "top-down",
			HarnessRecentDirections:    []string{"top-down", "bottom-up"},
		},
		Graph: mwsdbridge.GraphExport{
			SchemaVersion: mwsdbridge.GraphSchemaVersion,
			Root:          "/workspace",
			Documents: []mwsdbridge.Document{
				{
					Path:   "README.md",
					Links:  []string{"docs/GOAL.md"},
					Status: "",
				},
				{
					Path:   "docs/ROADMAP.md",
					ID:     "mws.roadmap",
					Title:  "로드맵",
					Status: "seed",
					Owner:  "local",
				},
				{
					Path:   "docs/GOAL.md",
					ID:     "mws.goal",
					Title:  "목표",
					Status: "seed",
					Owner:  "local",
				},
			},
			Stats: mwsdbridge.GraphStats{
				DocumentCount: 23,
				NodeCount:     23,
				EdgeCount:     100,
			},
		},
		Domain: mwsdbridge.DomainExport{
			SchemaVersion: mwsdbridge.DomainSchemaVersion,
			Path:          "/workspace/domains/macmini-workspace.lisp",
			Domain:        "macmini-workspace",
		},
		Harness: mwsdbridge.HarnessIndex{
			SchemaVersion:    mwsdbridge.HarnessSchemaVersion,
			Path:             "/workspace/harness/runs.jsonl",
			RunCount:         2,
			TopDownCount:     1,
			BottomUpCount:    1,
			LastDirection:    "bottom-up",
			NextDirection:    "top-down",
			RecentDirections: []string{"top-down", "bottom-up"},
		},
		Orchestration: mwsdbridge.OrchestrationSnapshot{
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
			ProviderCandidates: []mwsdbridge.ProviderCandidate{
				{ID: "codex", SourceWorkflow: "provider-selection", Available: true, ApprovalRequired: true},
				{ID: "claude-code", SourceWorkflow: "provider-selection", Available: true, ApprovalRequired: true},
				{ID: "cursor", SourceWorkflow: "provider-selection", Available: true, ApprovalRequired: true},
			},
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
		},
		Projects: mwsdbridge.ProjectRegistry{
			SchemaVersion:   mwsdbridge.ProjectsSchemaVersion,
			Root:            "/workspace",
			DomainPath:      "/workspace/domains/macmini-workspace.lisp",
			RepositoryCount: 3,
			Repositories: []mwsdbridge.ProjectRepository{
				{
					Name:          "macmini-workspace",
					Owner:         "kimjooyoon",
					Visibility:    "private",
					SSOTScope:     "workspace-control-plane",
					LocalPath:     "/Users/teddy/github/kimjooyoon/macmini-workspace",
					Remote:        "https://github.com/kimjooyoon/macmini-workspace",
					Role:          "control-plane",
					LocalPresent:  true,
					GitPresent:    true,
					RemoteMatches: true,
				},
				{
					Name:          "gui_engine",
					Owner:         "kimjooyoon",
					Visibility:    "private",
					SSOTScope:     "gui-engine",
					LocalPath:     "/Users/teddy/github/kimjooyoon/gui_engine",
					Remote:        "https://github.com/kimjooyoon/gui_engine",
					Role:          "gui-runtime",
					LocalPresent:  true,
					GitPresent:    true,
					RemoteMatches: true,
				},
				{
					Name:          "riido-daemon",
					Owner:         "kimjooyoon",
					Visibility:    "private",
					SSOTScope:     "project-daemon",
					LocalPath:     "/Users/teddy/github/kimjooyoon/riido-daemon",
					Remote:        "https://github.com/teamswyg/riido-daemon",
					Role:          "project-ssot",
					LocalPresent:  true,
					GitPresent:    true,
					RemoteMatches: true,
				},
			},
		},
	}
}
