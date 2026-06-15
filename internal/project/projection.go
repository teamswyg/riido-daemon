// Package project owns Riido's first project/workspace projection.
//
// It is intentionally downstream of mwsdbridge: macmini-workspace remains the
// local control-plane SSOT, while this package turns that snapshot into the
// shape Riido can use for project/task orchestration.
package project

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
