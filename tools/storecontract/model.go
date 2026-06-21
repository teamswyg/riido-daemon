package main

type contract struct {
	SchemaVersion            string    `json:"schema_version"`
	Product                  string    `json:"product"`
	Loop                     loopMeta  `json:"loop"`
	ProviderCLIBundling      string    `json:"provider_cli_bundling"`
	ExternalProviderCLINames []string  `json:"external_provider_cli_names"`
	StoreArtifactRoots       []string  `json:"store_artifact_roots"`
	RequiredDocs             []string  `json:"required_docs"`
	RequiredNoticeTerms      []string  `json:"required_notice_terms"`
	ChannelFiles             []string  `json:"channel_files,omitempty"`
	Channels                 []channel `json:"channels"`
}

type channel struct {
	ID                string   `json:"id"`
	LoopSource        string   `json:"loop_source"`
	Platform          string   `json:"platform"`
	Status            string   `json:"status"`
	RuntimeRole       string   `json:"runtime_role"`
	BackgroundRule    string   `json:"background_rule"`
	LocalIPCTransport string   `json:"local_ipc_transport"`
	DataRoot          string   `json:"data_root"`
	UpdateMechanism   string   `json:"update_mechanism"`
	RequiredSurfaces  []string `json:"required_surfaces"`
	ForbiddenSurfaces []string `json:"forbidden_surfaces"`
}

type checkResult struct {
	SchemaVersion      string           `json:"schema_version"`
	ContractPath       string           `json:"contract_path"`
	Product            string           `json:"product"`
	Status             string           `json:"status"`
	Channels           []string         `json:"channels"`
	StoreArtifactRoots []string         `json:"store_artifact_roots"`
	PolicyTablePath    string           `json:"policy_table_path,omitempty"`
	PolicyTableRows    []policyTableRow `json:"policy_table_rows,omitempty"`
	Errors             []string         `json:"errors,omitempty"`
}
