package agentbridge

// ToolRef identifies a single tool invocation within a run.
type ToolRef struct {
	ID                string
	Name              string
	Kind              string
	Args              map[string]string
	ProviderRequestID string
}
