package project

const StateSchemaVersion = "riido-project-state.v1"

// StateFile is Riido's first local persisted project/task view.
//
// It is a deterministic projection file, not the final transactional DB. The
// next storage layer can migrate this shape into SQLite or an event store
// without re-reading macmini-workspace files directly.
type StateFile struct {
	SchemaVersion          string                 `json:"schema_version"`
	ProjectionVersion      string                 `json:"projection_version"`
	Root                   string                 `json:"root"`
	Domain                 string                 `json:"domain"`
	HarnessRunCount        int                    `json:"harness_run_count"`
	HarnessNextDirection   string                 `json:"harness_next_direction"`
	OrchestrationMode      string                 `json:"orchestration_mode"`
	DecisionGate           string                 `json:"decision_gate"`
	DecisionBy             []string               `json:"decision_by"`
	DecisionLLMs           []string               `json:"decision_llms"`
	ProviderCandidates     []ProviderCandidate    `json:"provider_candidates"`
	RecommendedProvider    string                 `json:"recommended_provider"`
	RecommendedDecisionLLM string                 `json:"recommended_decision_llm"`
	NextAction             NextAction             `json:"next_action"`
	Projects               []ProjectState         `json:"projects"`
	Tasks                  []TaskState            `json:"tasks"`
	Diagnostics            []ProjectionDiagnostic `json:"diagnostics"`
}

type ProjectState struct {
	ID            string           `json:"id"`
	Owner         string           `json:"owner"`
	Visibility    string           `json:"visibility"`
	SSOTScope     string           `json:"ssot_scope"`
	LocalPath     string           `json:"local_path"`
	Remote        string           `json:"remote"`
	Role          string           `json:"role"`
	Health        RepositoryHealth `json:"health"`
	LocalPresent  bool             `json:"local_present"`
	GitPresent    bool             `json:"git_present"`
	RemoteMatches bool             `json:"remote_matches"`
}

type TaskState struct {
	ID                     string `json:"id"`
	ProjectID              string `json:"project_id"`
	State                  string `json:"state"`
	SourceDocumentID       string `json:"source_document_id"`
	SourceDocumentPath     string `json:"source_document_path"`
	Title                  string `json:"title"`
	Owner                  string `json:"owner"`
	SourceStatus           string `json:"source_status"`
	RecommendedProvider    string `json:"recommended_provider"`
	RecommendedDecisionLLM string `json:"recommended_decision_llm"`
	RequiresHumanApproval  bool   `json:"requires_human_approval"`
	HarnessNextDirection   string `json:"harness_next_direction"`
}
