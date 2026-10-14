package mwsdbridge

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
