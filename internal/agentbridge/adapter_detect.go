package agentbridge

// DetectEnv carries the environment an adapter consults during Detect.
// Adapters MUST read only from this struct, never os.Environ directly.
type DetectEnv struct {
	Executable  string
	PathExtra   []string
	EnvOverride map[string]string
}

// DetectResult is the snapshot of capability the adapter returns from Detect.
// RuntimeActor later reconciles this raw daemon observation into the C3
// ProviderCapability contract.
type DetectResult struct {
	Available         bool
	Executable        string
	Version           string
	SupportsStreaming bool
	SupportsResume    bool
	SupportsSystem    bool
	SupportsMaxTurns  bool
	SupportsMCP       bool
	SupportsToolHooks bool
	SupportsUsage     bool
	Reason            string
	Metadata          map[string]string
}
