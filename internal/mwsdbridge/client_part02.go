package mwsdbridge

import (
	"errors"
	"fmt"
)

// Validate checks the schema-level handshake between Riido and mwsd.
func (s Snapshot) Validate() error {
	checks := []struct {
		name string
		got  string
		want string
	}{
		{"graph", s.Graph.SchemaVersion, GraphSchemaVersion},
		{"domain", s.Domain.SchemaVersion, DomainSchemaVersion},
		{"harness", s.Harness.SchemaVersion, HarnessSchemaVersion},
		{"orchestration", s.Orchestration.SchemaVersion, OrchestrationSchemaVersion},
		{"projects", s.Projects.SchemaVersion, ProjectsSchemaVersion},
	}
	for _, check := range checks {
		if check.got != check.want {
			return fmt.Errorf("%s schema mismatch: got %q want %q", check.name, check.got, check.want)
		}
	}
	if s.Status.Root == "" {
		return errors.New("status root is empty")
	}
	if s.Graph.Root != "" && s.Graph.Root != s.Status.Root {
		return fmt.Errorf("graph root mismatch: %s != %s", s.Graph.Root, s.Status.Root)
	}
	if s.Status.OrchestrationSchemaVersion != "" && s.Status.OrchestrationSchemaVersion != OrchestrationSchemaVersion {
		return fmt.Errorf("status orchestration schema mismatch: got %q want %q", s.Status.OrchestrationSchemaVersion, OrchestrationSchemaVersion)
	}
	if s.Orchestration.Root != "" && s.Orchestration.Root != s.Status.Root {
		return fmt.Errorf("orchestration root mismatch: %s != %s", s.Orchestration.Root, s.Status.Root)
	}
	if s.Orchestration.DomainSchemaVersion != "" && s.Orchestration.DomainSchemaVersion != DomainSchemaVersion {
		return fmt.Errorf("orchestration domain schema mismatch: got %q want %q", s.Orchestration.DomainSchemaVersion, DomainSchemaVersion)
	}
	if s.Orchestration.HarnessSchemaVersion != "" && s.Orchestration.HarnessSchemaVersion != HarnessSchemaVersion {
		return fmt.Errorf("orchestration harness schema mismatch: got %q want %q", s.Orchestration.HarnessSchemaVersion, HarnessSchemaVersion)
	}
	if s.Orchestration.TopDownCount != s.Harness.TopDownCount {
		return fmt.Errorf("orchestration top-down count mismatch: %d != %d", s.Orchestration.TopDownCount, s.Harness.TopDownCount)
	}
	if s.Orchestration.BottomUpCount != s.Harness.BottomUpCount {
		return fmt.Errorf("orchestration bottom-up count mismatch: %d != %d", s.Orchestration.BottomUpCount, s.Harness.BottomUpCount)
	}
	if s.Orchestration.NextAction.Direction != "" && s.Orchestration.NextAction.Direction != s.Harness.NextDirection {
		return fmt.Errorf("orchestration next direction mismatch: %s != %s", s.Orchestration.NextAction.Direction, s.Harness.NextDirection)
	}
	if s.Projects.RepositoryCount != len(s.Projects.Repositories) {
		return fmt.Errorf("project registry count mismatch: %d != %d", s.Projects.RepositoryCount, len(s.Projects.Repositories))
	}
	return nil
}

type Status struct {
	Root                       string   `json:"root"`
	SocketPath                 string   `json:"socket_path"`
	GraphSchemaVersion         string   `json:"graph_schema_version"`
	DomainSchemaVersion        string   `json:"domain_schema_version"`
	HarnessSchemaVersion       string   `json:"harness_schema_version"`
	OrchestrationSchemaVersion string   `json:"orchestration_schema_version"`
	DocumentCount              int      `json:"document_count"`
	RepositoryCount            int      `json:"repository_count"`
	DomainName                 string   `json:"domain_name"`
	HarnessRunCount            int      `json:"harness_run_count"`
	HarnessNextDirection       string   `json:"harness_next_direction"`
	HarnessRecentDirections    []string `json:"harness_recent_directions"`
	SSOTConflictCount          int      `json:"ssot_conflict_count"`
	DomainDiagnosticCount      int      `json:"domain_diagnostic_count"`
	HarnessDiagnosticCount     int      `json:"harness_diagnostic_count"`
	DiagnosticCount            int      `json:"diagnostic_count"`
	ErrorCount                 int      `json:"error_count"`
	WarningCount               int      `json:"warning_count"`
	UnresolvedLinkCount        int      `json:"unresolved_link_count"`
}

type GraphExport struct {
	SchemaVersion string     `json:"schema_version"`
	Root          string     `json:"root"`
	Documents     []Document `json:"documents"`
	Stats         GraphStats `json:"stats"`
}

type Document struct {
	Path                string   `json:"path"`
	ID                  string   `json:"id"`
	Title               string   `json:"title"`
	Status              string   `json:"status"`
	Owner               string   `json:"owner"`
	Links               []string `json:"links"`
	Backlinks           []string `json:"backlinks"`
	MissingLinks        []string `json:"missing_links"`
	HasBacklinksSection bool     `json:"has_backlinks_section"`
}

type GraphStats struct {
	DocumentCount       int `json:"document_count"`
	NodeCount           int `json:"node_count"`
	EdgeCount           int `json:"edge_count"`
	DiagnosticCount     int `json:"diagnostic_count"`
	ErrorCount          int `json:"error_count"`
	WarningCount        int `json:"warning_count"`
	UnresolvedLinkCount int `json:"unresolved_link_count"`
}

type DomainExport struct {
	SchemaVersion string             `json:"schema_version"`
	Path          string             `json:"path"`
	Domain        string             `json:"domain"`
	Repositories  []DomainRepository `json:"repositories"`
	Diagnostics   []Diagnostic       `json:"diagnostics"`
}

type DomainRepository struct {
	Name       string   `json:"name"`
	Owner      string   `json:"owner"`
	Visibility string   `json:"visibility"`
	SSOTScope  string   `json:"ssot_scope"`
	LocalPath  string   `json:"local_path"`
	Remote     string   `json:"remote"`
	Role       string   `json:"role"`
	Consumes   []string `json:"consumes"`
}

type HarnessIndex struct {
	SchemaVersion             string       `json:"schema_version"`
	Path                      string       `json:"path"`
	RunCount                  int          `json:"run_count"`
	TopDownCount              int          `json:"top_down_count"`
	BottomUpCount             int          `json:"bottom_up_count"`
	LastDirection             string       `json:"last_direction"`
	NextDirection             string       `json:"next_direction"`
	ConsecutiveDirectionCount int          `json:"consecutive_direction_count"`
	RecentDirections          []string     `json:"recent_directions"`
	Diagnostics               []Diagnostic `json:"diagnostics"`
}

type OrchestrationSnapshot struct {
	SchemaVersion          string                  `json:"schema_version"`
	Root                   string                  `json:"root"`
	DomainPath             string                  `json:"domain_path"`
	HarnessRunPath         string                  `json:"harness_run_path"`
	DomainSchemaVersion    string                  `json:"domain_schema_version"`
	HarnessSchemaVersion   string                  `json:"harness_schema_version"`
	Mode                   string                  `json:"mode"`
	DecisionGate           string                  `json:"decision_gate"`
	DecisionBy             []string                `json:"decision_by"`
	DecisionLLMs           []string                `json:"decision_llms"`
	ProviderCandidates     []ProviderCandidate     `json:"provider_candidates"`
	RecommendedProvider    string                  `json:"recommended_provider"`
	RecommendedDecisionLLM string                  `json:"recommended_decision_llm"`
	NextAction             OrchestrationNextAction `json:"next_action"`
	TopDownCount           int                     `json:"top_down_count"`
	BottomUpCount          int                     `json:"bottom_up_count"`
	LastDirection          string                  `json:"last_direction"`
	Balanced               bool                    `json:"balanced"`
	DirectionBias          bool                    `json:"direction_bias"`
	Workflows              []OrchestrationWorkflow `json:"workflows"`
	RecentRuns             []OrchestrationRun      `json:"recent_runs"`
	Diagnostics            []Diagnostic            `json:"diagnostics"`
}

type ProviderCandidate struct {
	ID               string `json:"id"`
	SourceWorkflow   string `json:"source_workflow"`
	Available        bool   `json:"available"`
	ApprovalRequired bool   `json:"approval_required"`
}
