package agentbridge

import "time"

// NativeHistoryMessage is the v1 native_history_message metadata payload.
// It carries visible conversation text only; tool results and reasoning stay local.
type NativeHistoryMessage struct {
	SessionID  string    `json:"session_id"`
	MessageID  string    `json:"message_id"`
	Role       string    `json:"role"`
	Body       string    `json:"body"`
	ObservedAt time.Time `json:"observed_at,omitempty"`
	Ordinal    int       `json:"ordinal"`
}

const NativeHistoryMetadataKey = "native_history_message"
