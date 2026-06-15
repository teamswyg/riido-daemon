package main

import (
	"github.com/teamswyg/riido-daemon/internal/mwsdbridge"
)

func cliMwsdSnapshot() mwsdbridge.Snapshot {
	root := "/tmp/riido-cli-mwsd"
	return mwsdbridge.Snapshot{
		Status: mwsdbridge.Status{
			Root:                       root,
			SocketPath:                 "/tmp/mwsd.sock",
			GraphSchemaVersion:         mwsdbridge.GraphSchemaVersion,
			DomainSchemaVersion:        mwsdbridge.DomainSchemaVersion,
			HarnessSchemaVersion:       mwsdbridge.HarnessSchemaVersion,
			OrchestrationSchemaVersion: mwsdbridge.OrchestrationSchemaVersion,
			DocumentCount:              1,
			RepositoryCount:            1,
		},
		Graph: mwsdbridge.GraphExport{
			SchemaVersion: mwsdbridge.GraphSchemaVersion,
			Root:          root,
			Documents: []mwsdbridge.Document{{
				Path:   "docs/CLI.md",
				ID:     "mws.cli",
				Title:  "CLI migration",
				Status: "in-progress",
				Owner:  "kim",
			}},
			Stats: mwsdbridge.GraphStats{
				DocumentCount: 1,
				NodeCount:     1,
				EdgeCount:     0,
			},
		},
		Domain: mwsdbridge.DomainExport{
			SchemaVersion: mwsdbridge.DomainSchemaVersion,
			Path:          "docs/domain.mws",
			Domain:        "macmini-workspace",
		},
		Harness: mwsdbridge.HarnessIndex{
			SchemaVersion:    mwsdbridge.HarnessSchemaVersion,
			RunCount:         1,
			TopDownCount:     1,
			BottomUpCount:    0,
			LastDirection:    "top-down",
			NextDirection:    "bottom-up",
			RecentDirections: []string{"top-down"},
		},
		Orchestration: mwsdbridge.OrchestrationSnapshot{
			SchemaVersion:          mwsdbridge.OrchestrationSchemaVersion,
			Root:                   root,
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
		},
		Projects: mwsdbridge.ProjectRegistry{
			SchemaVersion:   mwsdbridge.ProjectsSchemaVersion,
			Root:            root,
			RepositoryCount: 1,
			Repositories: []mwsdbridge.ProjectRepository{{
				Name:          "riido-daemon",
				Owner:         "teamswyg",
				Visibility:    "private",
				SSOTScope:     "docs",
				LocalPath:     root,
				Remote:        "https://github.com/teamswyg/riido-daemon",
				Role:          "daemon",
				LocalPresent:  true,
				GitPresent:    true,
				RemoteMatches: true,
			}},
		},
	}
}
