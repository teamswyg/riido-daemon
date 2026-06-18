// Package project owns Riido's first project/workspace projection.
//
// It is intentionally downstream of mwsdbridge: macmini-workspace remains the
// local control-plane SSOT, while this package turns that snapshot into the
// shape Riido can use for project/task orchestration.
package project

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
