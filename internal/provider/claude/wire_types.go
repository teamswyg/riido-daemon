package claude

type wireEventType string

const (
	wireEventMalformed    wireEventType = "malformed"
	wireEventSystem       wireEventType = "system"
	wireEventAssistant    wireEventType = "assistant"
	wireEventUser         wireEventType = "user"
	wireEventControl      wireEventType = "control_request"
	wireEventResult       wireEventType = "result"
	wireEventLog          wireEventType = "log"
	wireEventError        wireEventType = "error"
	wireEventRateLimit    wireEventType = "rate_limit"
	wireEventRateLimitAlt wireEventType = "rate_limit_event"
)

type wireContentType string

const (
	wireContentText       wireContentType = "text"
	wireContentThinking   wireContentType = "thinking"
	wireContentToolUse    wireContentType = "tool_use"
	wireContentToolResult wireContentType = "tool_result"
)

type wireControlSubtype string

const wireControlPermissionRequest wireControlSubtype = "permission_request"

type wireResultSubtype string

const (
	wireResultSubtypeError          wireResultSubtype = "error"
	wireResultSubtypeExecutionError wireResultSubtype = "error_during_execution"
	wireResultSubtypeCancelled      wireResultSubtype = "cancelled"
	wireResultSubtypeMaxTurns       wireResultSubtype = "max_turns"
)
