package workdir

// ProviderNativeConfigPlan is the deterministic file plan for one provider's
// native config materialization.
type ProviderNativeConfigPlan struct {
	ProviderKind           string
	PrimaryInstructionFile string
	ManifestFile           string
	HookMode               string
	ConfigHomeDir          string
	ProviderSettingsFiles  []string
	HookFiles              []string
	ExtraFiles             []string
}

// GeneratedFiles returns the relative file paths materialized by the plan.
func (p ProviderNativeConfigPlan) GeneratedFiles() []string {
	files := []string{p.PrimaryInstructionFile, p.ManifestFile}
	files = append(files, p.ProviderSettingsFiles...)
	files = append(files, p.HookFiles...)
	files = append(files, p.ExtraFiles...)
	return sortedUniquePaths(files)
}

// NativeConfigManifest is the replayable evidence of what C6 materialized
// into a run workdir and its native-config copy.
type NativeConfigManifest struct {
	SchemaVersion              string   `json:"schema_version"`
	ProviderKind               string   `json:"provider_kind"`
	ProtocolKind               string   `json:"protocol_kind,omitempty"`
	PrimaryInstructionFile     string   `json:"primary_instruction_file"`
	ManifestFile               string   `json:"manifest_file"`
	HookMode                   string   `json:"hook_mode"`
	ConfigHomeDir              string   `json:"config_home_dir,omitempty"`
	ProviderSettingsFiles      []string `json:"provider_settings_files,omitempty"`
	HookFiles                  []string `json:"hook_files,omitempty"`
	TelemetryContractPlacement string   `json:"telemetry_contract_placement,omitempty"`
	Workflow                   string   `json:"workflow"`
	GeneratedFiles             []string `json:"generated_files"`
}

// NativeConfigVersionInput is the C6 execution-context fingerprint input.
type NativeConfigVersionInput struct {
	PolicyBundleVersion string
	ProviderKind        string
	ProtocolKind        string
}

type nativeConfigVersionDoc struct {
	PolicyBundleVersion string           `json:"policyBundleVersion"`
	NativeConfigPlan    nativeConfigPlan `json:"nativeConfigPlan"`
	SchemaVersion       int              `json:"schemaVersion"`
	Unknown             map[string]any   `json:"unknown,omitempty"`
}

type nativeConfigPlan struct {
	ProviderKind       string                 `json:"providerKind"`
	ProtocolKind       string                 `json:"protocolKind"`
	InjectedFiles      []nativeConfigFileHash `json:"injectedFiles"`
	HookScriptVersions []nativeConfigFileHash `json:"hookScriptVersions,omitempty"`
	WrapperManifestSHA string                 `json:"wrapperManifestSha,omitempty"`
}

type nativeConfigFileHash struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}
