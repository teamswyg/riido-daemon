package cursor

type wireEventType string

const (
	wireEventMalformed  wireEventType = "malformed"
	wireEventSystem     wireEventType = "system"
	wireEventText       wireEventType = "text"
	wireEventAssistant  wireEventType = "assistant"
	wireEventToolUse    wireEventType = "tool_use"
	wireEventToolResult wireEventType = "tool_result"
	wireEventResult     wireEventType = "result"
	wireEventStepFinish wireEventType = "step_finish"
	wireEventError      wireEventType = "error"
)

type wireContentType string

const (
	wireContentText       wireContentType = "text"
	wireContentOutputText wireContentType = "output_text"
	wireContentThinking   wireContentType = "thinking"
	wireContentToolUse    wireContentType = "tool_use"
)

type wireResultSubtype string

const (
	wireResultSubtypeError          wireResultSubtype = "error"
	wireResultSubtypeExecutionError wireResultSubtype = "error_during_execution"
	wireResultSubtypeCancelled      wireResultSubtype = "cancelled"
)
