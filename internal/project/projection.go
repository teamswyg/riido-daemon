// Package project owns Riido's first project/workspace projection.
//
// It is intentionally downstream of mwsdbridge: macmini-workspace remains the
// local control-plane SSOT, while this package turns that snapshot into the
// shape Riido can use for project/task orchestration.
package project

import (
	"fmt"
	"sort"

	"github.com/teamswyg/riido-daemon/internal/mwsdbridge"
)

type RepositoryHealth string

const (
	RepositoryReady          RepositoryHealth = "ready"
	RepositoryMissingLocal   RepositoryHealth = "missing-local"
	RepositoryMissingGit     RepositoryHealth = "missing-git"
	RepositoryRemoteMismatch RepositoryHealth = "remote-mismatch"
)

// WorkspaceProjection is a deterministic view of the Mac mini control-plane
// state that Riido can consume without knowing mwsd's raw response envelopes.
type WorkspaceProjection struct {
	SchemaVersion           string                 `json:"schema_version"`
	Root                    string                 `json:"root"`
	Domain                  string                 `json:"domain"`
	DomainPath              string                 `json:"domain_path"`
	DocumentCount           int                    `json:"document_count"`
	GraphNodeCount          int                    `json:"graph_node_count"`
	GraphEdgeCount          int                    `json:"graph_edge_count"`
	HarnessRunCount         int                    `json:"harness_run_count"`
	HarnessNextDirection    string                 `json:"harness_next_direction"`
	HarnessRecentDirections []string               `json:"harness_recent_directions"`
	OrchestrationSchema     string                 `json:"orchestration_schema_version"`
	OrchestrationMode       string                 `json:"orchestration_mode"`
	DecisionGate            string                 `json:"decision_gate"`
	DecisionBy              []string               `json:"decision_by"`
	DecisionLLMs            []string               `json:"decision_llms"`
	ProviderCandidates      []ProviderCandidate    `json:"provider_candidates"`
	RecommendedProvider     string                 `json:"recommended_provider"`
	RecommendedDecisionLLM  string                 `json:"recommended_decision_llm"`
	NextAction              NextAction             `json:"next_action"`
	HarnessBalanced         bool                   `json:"harness_balanced"`
	DirectionBias           bool                   `json:"direction_bias"`
	RecentProviderRuns      []ProviderRunSummary   `json:"recent_provider_runs"`
	SSOTConflictCount       int                    `json:"ssot_conflict_count"`
	Projects                []Project              `json:"projects"`
	DocumentTaskLinks       []DocumentTaskLink     `json:"document_task_links"`
	Diagnostics             []ProjectionDiagnostic `json:"diagnostics"`
}

func (p WorkspaceProjection) Ready() bool {
	for _, diagnostic := range p.Diagnostics {
		if diagnostic.Severity == "error" {
			return false
		}
	}
	return true
}

type Project struct {
	ID            string           `json:"id"`
	Owner         string           `json:"owner"`
	Visibility    string           `json:"visibility"`
	SSOTScope     string           `json:"ssot_scope"`
	LocalPath     string           `json:"local_path"`
	Remote        string           `json:"remote"`
	Role          string           `json:"role"`
	Consumes      []string         `json:"consumes"`
	Health        RepositoryHealth `json:"health"`
	LocalPresent  bool             `json:"local_present"`
	GitPresent    bool             `json:"git_present"`
	RemoteMatches bool             `json:"remote_matches"`
}

type ProviderCandidate struct {
	ID               string `json:"id"`
	SourceWorkflow   string `json:"source_workflow"`
	Available        bool   `json:"available"`
	ApprovalRequired bool   `json:"approval_required"`
}

type NextAction struct {
	Direction             string `json:"direction"`
	CommandSurface        string `json:"command_surface"`
	Reason                string `json:"reason"`
	RequiresHumanApproval bool   `json:"requires_human_approval"`
}

type ProviderRunSummary struct {
	ID        string `json:"id"`
	Direction string `json:"direction"`
	Source    string `json:"source"`
	Provider  string `json:"provider"`
	Command   string `json:"command"`
	Result    string `json:"result"`
}

// DocumentTaskLink is the first deterministic bridge from document SSOT to
// Riido task identity. It is not yet a persisted task row; it is the stable
// source mapping the future task store can consume.
type DocumentTaskLink struct {
	TaskID                 string `json:"task_id"`
	DocumentID             string `json:"document_id"`
	DocumentPath           string `json:"document_path"`
	Title                  string `json:"title"`
	Status                 string `json:"status"`
	Owner                  string `json:"owner"`
	ProjectID              string `json:"project_id"`
	RecommendedProvider    string `json:"recommended_provider"`
	RecommendedDecisionLLM string `json:"recommended_decision_llm"`
	RequiresHumanApproval  bool   `json:"requires_human_approval"`
	HarnessNextDirection   string `json:"harness_next_direction"`
}

type ProjectionDiagnostic struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}

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

func providerRuns(runs []mwsdbridge.OrchestrationRun) []ProviderRunSummary {
	out := make([]ProviderRunSummary, 0, len(runs))
	for _, run := range runs {
		out = append(out, ProviderRunSummary{
			ID:        run.ID,
			Direction: run.Direction,
			Source:    run.Source,
			Provider:  run.Provider,
			Command:   run.Command,
			Result:    run.Result,
		})
	}
	return out
}

func hasProject(projects []Project, id string) bool {
	for _, project := range projects {
		if project.ID == id {
			return true
		}
	}
	return false
}

func repositoryHealth(repository mwsdbridge.ProjectRepository) RepositoryHealth {
	switch {
	case !repository.LocalPresent:
		return RepositoryMissingLocal
	case !repository.GitPresent:
		return RepositoryMissingGit
	case !repository.RemoteMatches:
		return RepositoryRemoteMismatch
	default:
		return RepositoryReady
	}
}

func liftDiagnostics(source string, diagnostics []mwsdbridge.Diagnostic) []ProjectionDiagnostic {
	out := make([]ProjectionDiagnostic, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		out = append(out, ProjectionDiagnostic{
			Severity: diagnostic.Severity,
			Code:     source + "-" + diagnostic.Code,
			Message:  diagnostic.Message,
		})
	}
	return out
}
