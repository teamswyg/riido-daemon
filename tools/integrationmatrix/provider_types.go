package main

type providerValidation struct {
	SchemaVersion   string             `json:"schema_version"`
	ID              string             `json:"id"`
	RiidoTask       string             `json:"riido_task"`
	HumanDoc        string             `json:"human_doc"`
	SourceDocuments []string           `json:"source_documents"`
	GlobalRules     []string           `json:"global_rules"`
	ProviderFiles   []string           `json:"provider_files"`
	Providers       []providerEvidence `json:"providers"`
}

type providerEvidence struct {
	Provider                  string   `json:"provider"`
	DisplayName               string   `json:"display_name"`
	RuntimeKind               string   `json:"runtime_kind"`
	Executable                string   `json:"executable"`
	DeterministicCI           []string `json:"deterministic_ci"`
	OptInIntegration          string   `json:"opt_in_integration"`
	WorktreeSupport           string   `json:"worktree_support"`
	PassEvidence              []string `json:"pass_evidence"`
	NegativeOrLimitedEvidence []string `json:"negative_or_limited_evidence"`
	RequiredSchedulingGate    []string `json:"required_scheduling_gate"`
	SkipBeforeRun             []string `json:"skip_before_run"`
	LatestEvidence            []string `json:"latest_evidence"`
	MustNotClaim              []string `json:"must_not_claim"`
}
