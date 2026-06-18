package mwsdbridge

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
