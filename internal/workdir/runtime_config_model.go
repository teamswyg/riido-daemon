package workdir

// RuntimeConfig is the inputs the workdir adapter renders into the
// provider's native config file (per spec §10 Phase 7, §7).
//
// The four sections mirror Multica's native-config 4단 structure
// (multica.md §7).
type RuntimeConfig struct {
	Provider                   string   // e.g. "claude", "codex"
	ProtocolKind               string   // C3 protocol selected for the run
	TelemetryContractPlacement string   // prompt/system-prompt/... when injected
	NativeHookMode             string   // C7 decision result; empty uses the provider plan default
	NativeConfigHomeMode       string   // C7 decision result; empty uses the provider plan default
	Identity                   string   // "You are: <agent name> (id: ...)"
	CLICatalog                 []string // command examples
	HardRules                  []string // invariants the agent must follow
	Workflow                   string   // workflow branch label (chat|quick-create|...)
}
