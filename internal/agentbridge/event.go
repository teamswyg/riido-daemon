package agentbridge

import "time"

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
// Stderr/log spam and process-level signals do NOT count — only events
// that show the provider is making progress on the task.
// Approval requests count because the session actor owns the wait timeout
// after a provider-native ApprovalRequested event.
//
// See docs/20-domain/provider-runtime.md §5.5 and §5.6.
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

// Event is the run-scope envelope passed to the reducer. Fields not
// applicable to a given Kind are simply zero-valued.
type Event struct {
	Kind      EventKind
	At        time.Time
	SessionID string
	Phase     RunState
	Text      string
	Tool      ToolRef
	Usage     Usage
	Result    Result
	ExitCode  int
	Err       string
}

// ToolRef identifies a single tool invocation within a run.
type ToolRef struct {
	ID                string
	Name              string
	Kind              string
	Args              map[string]string
	ProviderRequestID string
}

// Usage is the provider-neutral token-usage accumulator. Each provider
// reports usage in its own schema; adapters normalize into this struct
// (docs/20-domain/provider-runtime.md §5.5).
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	ReasoningTokens  int
	CacheReadTokens  int
	CacheWriteTokens int
}

// Add returns the field-wise sum of u and other.
func (u Usage) Add(other Usage) Usage {
	return Usage{
		PromptTokens:     u.PromptTokens + other.PromptTokens,
		CompletionTokens: u.CompletionTokens + other.CompletionTokens,
		ReasoningTokens:  u.ReasoningTokens + other.ReasoningTokens,
		CacheReadTokens:  u.CacheReadTokens + other.CacheReadTokens,
		CacheWriteTokens: u.CacheWriteTokens + other.CacheWriteTokens,
	}
}
