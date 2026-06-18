package agentbridge

import "context"

// Adapter is the run-scope provider port: a single provider CLI plugin.
//
// Adapters are translators, not state owners. They build a process command,
// parse raw provider output, and translate each raw provider event into one or
// more run-scope Events.
type Adapter interface {
	Name() string
	Detect(ctx context.Context, env DetectEnv) (DetectResult, error)
	BuildStart(req StartRequest) (StartCommand, error)
	NewParser() Parser
	Translate(raw RawEvent) ([]Event, []Command, error)
	BlockedArgs() []string
}
