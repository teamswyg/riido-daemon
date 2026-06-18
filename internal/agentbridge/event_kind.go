package agentbridge

// EventKind is the run-scope event taxonomy. These events are produced
// by an Adapter's Translate method from raw provider output and consumed
// by the reducer.
type EventKind string

const (
	EventLifecycle          EventKind = "lifecycle"
	EventSessionIdentified  EventKind = "session_identified"
	EventTextDelta          EventKind = "text_delta"
	EventThinkingDelta      EventKind = "thinking_delta"
	EventToolCallStarted    EventKind = "tool_call_started"
	EventToolCallDelta      EventKind = "tool_call_delta"
	EventToolCallCompleted  EventKind = "tool_call_completed"
	EventToolCallFailed     EventKind = "tool_call_failed"
	EventToolApprovalNeeded EventKind = "tool_approval_needed"
	EventUsageDelta         EventKind = "usage_delta"
	EventProgress           EventKind = "progress"
	EventLog                EventKind = "log"
	EventWarning            EventKind = "warning"
	EventError              EventKind = "error"
	EventResult             EventKind = "result"
	EventCancellation       EventKind = "cancellation_requested"
	EventTimeout            EventKind = "timeout"
	EventProcessExit        EventKind = "process_exit"
)

// IsSemanticActivity reports whether an event resets the idle watchdog.
// Stderr/log spam and process-level signals do not count; only events that
// show provider task progress reset it.
func (k EventKind) IsSemanticActivity() bool {
	switch k {
	case EventLifecycle,
		EventTextDelta, EventThinkingDelta,
		EventToolCallStarted, EventToolCallDelta,
		EventToolCallCompleted, EventToolCallFailed,
		EventToolApprovalNeeded,
		EventUsageDelta,
		EventProgress:
		return true
	default:
		return false
	}
}
