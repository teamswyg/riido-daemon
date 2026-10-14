package saasplane

// AgentBinding maps a SaaS agent identity to one local provider runtime.
type AgentBinding struct {
	AgentID         string
	RuntimeProvider string
}
