package project

import (
	"fmt"
	"sort"

	"github.com/teamswyg/riido-daemon/internal/mwsdbridge"
)

func FromMwsdSnapshot(snapshot mwsdbridge.Snapshot) (WorkspaceProjection, error) {
	if err := snapshot.Validate(); err != nil {
		return WorkspaceProjection{}, err
	}

	projection := WorkspaceProjection{
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

	for _, repository := range snapshot.Projects.Repositories {
		project := Project{
			ID:            repository.Name,
			Owner:         repository.Owner,
			Visibility:    repository.Visibility,
			SSOTScope:     repository.SSOTScope,
			LocalPath:     repository.LocalPath,
			Remote:        repository.Remote,
			Role:          repository.Role,
			Consumes:      append([]string(nil), repository.Consumes...),
			Health:        repositoryHealth(repository),
			LocalPresent:  repository.LocalPresent,
			GitPresent:    repository.GitPresent,
			RemoteMatches: repository.RemoteMatches,
		}
		projection.Projects = append(projection.Projects, project)
		if project.Visibility != "private" {
			projection.Diagnostics = append(projection.Diagnostics, ProjectionDiagnostic{
				Severity: "warning",
				Code:     "project-not-private",
				Message:  fmt.Sprintf("project %s visibility is %s", project.ID, project.Visibility),
			})
		}
		if project.Health != RepositoryReady {
			projection.Diagnostics = append(projection.Diagnostics, ProjectionDiagnostic{
				Severity: "error",
				Code:     "project-repository-not-ready",
				Message:  fmt.Sprintf("project %s repository health is %s", project.ID, project.Health),
			})
		}
	}

	sort.Slice(projection.Projects, func(i, j int) bool {
		return projection.Projects[i].ID < projection.Projects[j].ID
	})
	projection.DocumentTaskLinks = documentTaskLinks(snapshot.Graph.Documents, projection)
	projection.Diagnostics = append(projection.Diagnostics, liftDiagnostics("domain", snapshot.Domain.Diagnostics)...)
	projection.Diagnostics = append(projection.Diagnostics, liftDiagnostics("harness", snapshot.Harness.Diagnostics)...)
	projection.Diagnostics = append(projection.Diagnostics, liftDiagnostics("orchestration", snapshot.Orchestration.Diagnostics)...)
	projection.Diagnostics = append(projection.Diagnostics, liftDiagnostics("projects", snapshot.Projects.Diagnostics)...)
	if projection.DecisionGate != "human-approval-required" {
		projection.Diagnostics = append(projection.Diagnostics, ProjectionDiagnostic{
			Severity: "error",
			Code:     "orchestration-human-gate-missing",
			Message:  fmt.Sprintf("orchestration decision gate is %s", projection.DecisionGate),
		})
	}
	if projection.RecommendedProvider == "" {
		projection.Diagnostics = append(projection.Diagnostics, ProjectionDiagnostic{
			Severity: "error",
			Code:     "orchestration-recommended-provider-missing",
			Message:  "orchestration has no recommended provider",
		})
	}
	if projection.RecommendedDecisionLLM == "" {
		projection.Diagnostics = append(projection.Diagnostics, ProjectionDiagnostic{
			Severity: "error",
			Code:     "orchestration-recommended-decision-llm-missing",
			Message:  "orchestration has no recommended decision LLM",
		})
	}
	if projection.DirectionBias {
		projection.Diagnostics = append(projection.Diagnostics, ProjectionDiagnostic{
			Severity: "warning",
			Code:     "orchestration-direction-biased",
			Message:  "orchestration reports a top-down/bottom-up direction bias",
		})
	}
	if snapshot.Status.SSOTConflictCount > 0 {
		projection.Diagnostics = append(projection.Diagnostics, ProjectionDiagnostic{
			Severity: "error",
			Code:     "ssot-conflicts-present",
			Message:  fmt.Sprintf("mwsd status reports %d SSOT conflicts", snapshot.Status.SSOTConflictCount),
		})
	}
	if projection.Diagnostics == nil {
		projection.Diagnostics = []ProjectionDiagnostic{}
	}
	return projection, nil
}

func documentTaskLinks(documents []mwsdbridge.Document, projection WorkspaceProjection) []DocumentTaskLink {
	projectID := "macmini-workspace"
	if !hasProject(projection.Projects, projectID) && len(projection.Projects) > 0 {
		projectID = projection.Projects[0].ID
	}
	links := make([]DocumentTaskLink, 0, len(documents))
	for _, document := range documents {
		if document.ID == "" {
			continue
		}
		links = append(links, DocumentTaskLink{
			TaskID:                 "task:" + document.ID,
			DocumentID:             document.ID,
			DocumentPath:           document.Path,
			Title:                  document.Title,
			Status:                 document.Status,
			Owner:                  document.Owner,
			ProjectID:              projectID,
			RecommendedProvider:    projection.RecommendedProvider,
			RecommendedDecisionLLM: projection.RecommendedDecisionLLM,
			RequiresHumanApproval:  projection.DecisionGate == "human-approval-required" || projection.NextAction.RequiresHumanApproval,
			HarnessNextDirection:   projection.HarnessNextDirection,
		})
	}
	sort.Slice(links, func(i, j int) bool {
		return links[i].DocumentID < links[j].DocumentID
	})
	return links
}

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
