package main

type nativeConfigPlanCatalogSpec struct {
	SchemaVersion         string                         `json:"schema_version"`
	ManifestSchemaVersion string                         `json:"manifest_schema_version"`
	Default               nativeConfigProviderPlanSpec   `json:"default"`
	Providers             []nativeConfigProviderPlanSpec `json:"providers"`
	SpecPath              string
}

type nativeConfigProviderPlanSpec struct {
	ProviderKind           string   `json:"provider_kind,omitempty"`
	PrimaryInstructionFile string   `json:"primary_instruction_file"`
	ManifestFile           string   `json:"manifest_file"`
	HookMode               string   `json:"hook_mode"`
	ConfigHomeDir          string   `json:"config_home_dir,omitempty"`
	ProviderSettingsFiles  []string `json:"provider_settings_files,omitempty"`
	HookFiles              []string `json:"hook_files,omitempty"`
	ExtraFiles             []string `json:"extra_files,omitempty"`
}
