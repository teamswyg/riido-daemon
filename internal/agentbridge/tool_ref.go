package agentbridge

// ToolRef identifies a single tool invocation within a run.
type ToolRef struct {
	ID                string
	Name              string
	Kind              string
	Args              map[string]string
	ProviderRequestID string
	RequireHuman      bool
	// ContinueAfterRejection is set only by adapters whose native rejection
	// response denies this tool without terminating the owned turn.
	ContinueAfterRejection bool
}
