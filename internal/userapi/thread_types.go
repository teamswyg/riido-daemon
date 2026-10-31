package userapi

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var ErrThreadNotRetained = errors.New("the requested thread is not in the currently visible retained history")

type ThreadReadClient interface {
	ReadTaskThreads(context.Context, string, string) ([]byte, int, error)
}
type ThreadReader struct{ Client ThreadReadClient }

type ThreadMessage struct {
	DisplayMessage    string    `json:"display_message,omitempty"`
	MessageID         string    `json:"message_id"`
	Role              string    `json:"role"`
	CommentKind       string    `json:"comment_kind,omitempty"`
	AssignmentID      string    `json:"assignment_id,omitempty"`
	RunID             string    `json:"run_id,omitempty"`
	SourceMessageID   string    `json:"source_message_id,omitempty"`
	AuthorPrincipalID string    `json:"author_principal_id,omitempty"`
	Seq               int       `json:"seq,omitempty"`
	Body              string    `json:"body,omitempty"`
	ResultMessage     string    `json:"result_message,omitempty"`
	ObservedAt        time.Time `json:"observed_at,omitzero"`
}

type ThreadStream struct {
	Rel       string `json:"rel"`
	Href      string `json:"href"`
	EventType string `json:"event_type"`
	TaskID    string `json:"task_id"`
	ThreadID  string `json:"thread_id"`
	RunID     string `json:"run_id"`
}
type ThreadFailure struct {
	ResultStatus    string `json:"result_status,omitempty"`
	FailureCategory string `json:"failure_category,omitempty"`
	Message         string `json:"message,omitempty"`
}
type ThreadIdentity struct {
	ThreadID        string    `json:"thread_id"`
	ConversationID  string    `json:"conversation_id"`
	ParentThreadID  string    `json:"parent_thread_id,omitempty"`
	TaskID          string    `json:"task_id"`
	AssignmentID    string    `json:"assignment_id,omitempty"`
	AgentID         string    `json:"agent_id"`
	AgentSnapshotID string    `json:"agent_snapshot_id,omitempty"`
	RunID           string    `json:"run_id"`
	WorkStatus      string    `json:"work_status"`
	AssignmentState string    `json:"assignment_state"`
	StartedAt       time.Time `json:"started_at,omitzero"`
	CompletedAt     time.Time `json:"completed_at,omitzero"`
}
type ThreadRecord struct {
	ThreadIdentity
	ExecutionContext   json.RawMessage `json:"execution_context,omitempty"`
	FailureDiagnostics *ThreadFailure  `json:"failure_diagnostics,omitempty"`
	Messages           []ThreadMessage `json:"messages"`
	ActiveStream       *ThreadStream   `json:"active_stream,omitempty"`
}
type ThreadSummary struct {
	ThreadIdentity
	MessageCount int `json:"message_count"`
}
type ThreadAgentSnapshot struct {
	AgentID             string    `json:"agent_id"`
	WorkspaceID         string    `json:"workspace_id,omitempty"`
	OwnerPrincipalID    string    `json:"owner_principal_id,omitempty"`
	Name                string    `json:"name,omitempty"`
	ProfileThumbnailURL string    `json:"profile_thumbnail_url,omitempty"`
	TmpColor            string    `json:"tmp_color,omitempty"`
	ProfileColor        string    `json:"profile_color,omitempty"`
	Visibility          string    `json:"visibility,omitempty"`
	RuntimeKind         string    `json:"runtime_kind,omitempty"`
	ModelID             string    `json:"model_id,omitempty"`
	ModelLabel          string    `json:"model_label,omitempty"`
	CapturedAt          time.Time `json:"captured_at,omitzero"`
}
type RetainedHistory struct {
	SchemaVersion      string `json:"schema_version"`
	WorkspaceID        string `json:"workspace_id"`
	TaskID             string `json:"task_id"`
	HistoryScope       string `json:"history_scope"`
	ProjectionComplete bool   `json:"projection_complete"`
	Complete           bool   `json:"complete"`
}
type ThreadList struct {
	RetainedHistory
	AssignedAgentIDs []string                        `json:"assigned_agent_ids"`
	Threads          []ThreadSummary                 `json:"threads"`
	AgentSnapshots   map[string]*ThreadAgentSnapshot `json:"agent_snapshots,omitempty"`
}
type ThreadShow struct {
	RetainedHistory
	Thread        ThreadRecord         `json:"thread"`
	AgentSnapshot *ThreadAgentSnapshot `json:"agent_snapshot,omitempty"`
}
