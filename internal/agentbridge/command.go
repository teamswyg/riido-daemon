package agentbridge

// CommandKind enumerates the imperative outputs the reducer can return
// alongside a state transition. Commands are executed by the session
// actor (not the reducer); the reducer remains a pure function.
type CommandKind string

const (
	CommandStartProvider      CommandKind = "start_provider"
	CommandWriteProviderInput CommandKind = "write_provider_input"
	CommandApproveTool        CommandKind = "approve_tool"
	CommandRejectTool         CommandKind = "reject_tool"
	CommandCancelProvider     CommandKind = "cancel_provider"
	CommandPersistSession     CommandKind = "persist_session"
	CommandFlushEvents        CommandKind = "flush_events"
)

// Command is the reducer's imperative output. ToolID is meaningful only
// for CommandApproveTool / CommandRejectTool. ProviderRequestID carries
// provider transport correlation ids such as Claude control_request.request_id.
type Command struct {
	Kind              CommandKind
	ToolID            string
	ProviderRequestID string
	Input             []byte
	Reason            string
}

// ProviderInputBuilder is an optional adapter-side ACL that translates
// provider-neutral reducer commands into bytes for the provider stdin protocol.
// Adapters that do not have a long-lived stdin control protocol can ignore it.
type ProviderInputBuilder interface {
	BuildProviderInput(Command) ([]byte, error)
}
