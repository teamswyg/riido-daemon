package runtimeactor

// Capability is the daemon-side view of a single provider's runtime
// readiness. Built from Adapter.Detect.
type Capability struct {
	Provider                  string `json:"provider"`
	Available                 bool   `json:"available"`
	Version                   string `json:"version,omitempty"`
	Executable                string `json:"executable,omitempty"`
	Profile                   string `json:"profile,omitempty"`
	Reason                    string `json:"reason,omitempty"`
	ProtocolKind              string `json:"protocol_kind,omitempty"`
	AdapterID                 string `json:"adapter_id,omitempty"`
	AdapterVersion            string `json:"adapter_version,omitempty"`
	ProtocolVersion           string `json:"protocol_version,omitempty"`
	CompatibilityStatus       string `json:"compatibility_status,omitempty"`
	CapabilityFingerprint     string `json:"capability_fingerprint,omitempty"`
	DetectedFingerprint       string `json:"detected_fingerprint,omitempty"`
	RequiresExperimentalOptIn bool   `json:"requires_experimental_opt_in,omitempty"`
	SupportsStreaming         bool   `json:"supports_streaming"`
	SupportsResume            bool   `json:"supports_resume"`
	SupportsSystem            bool   `json:"supports_system"`
	SupportsMaxTurns          bool   `json:"supports_max_turns"`
	SupportsMCP               bool   `json:"supports_mcp"`
	SupportsToolHooks         bool   `json:"supports_tool_hooks"`
	SupportsUsage             bool   `json:"supports_usage"`
	SupportsFileEvents        bool   `json:"supports_file_events"`
	SupportsWorktree          bool   `json:"supports_worktree"`
}
