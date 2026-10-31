package userapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"unicode/utf8"
)

type threadCollection struct {
	SchemaVersion    string                          `json:"schema_version"`
	TaskID           string                          `json:"task_id"`
	AssignedAgentIDs []string                        `json:"assigned_agent_ids"`
	Threads          []ThreadRecord                  `json:"threads"`
	AgentSnapshots   map[string]*ThreadAgentSnapshot `json:"agent_snapshots,omitempty"`
	ActiveStream     *ThreadStream                   `json:"active_stream,omitempty"`
}

func (r ThreadReader) List(ctx context.Context, workspace, task string) (*ThreadList, error) {
	collection, err := r.read(ctx, workspace, task)
	if err != nil {
		return nil, err
	}
	result := &ThreadList{RetainedHistory: retainedHistory(workspace, task), AssignedAgentIDs: collection.AssignedAgentIDs, AgentSnapshots: collection.AgentSnapshots, Threads: make([]ThreadSummary, 0, len(collection.Threads))}
	for _, thread := range collection.Threads {
		result.Threads = append(result.Threads, ThreadSummary{ThreadIdentity: thread.ThreadIdentity, MessageCount: len(thread.Messages)})
	}
	return result, nil
}

func (r ThreadReader) Show(ctx context.Context, workspace, task, id string) (*ThreadShow, error) {
	if !ValidThreadReadID(id) {
		return nil, ErrInvalidRequest
	}
	collection, err := r.read(ctx, workspace, task)
	if err != nil {
		return nil, err
	}
	for _, thread := range collection.Threads {
		if thread.ThreadID == id {
			return &ThreadShow{RetainedHistory: retainedHistory(workspace, task), Thread: thread, AgentSnapshot: collection.AgentSnapshots[thread.AgentSnapshotID]}, nil
		}
	}
	return nil, ErrThreadNotRetained
}

func retainedHistory(workspace, task string) RetainedHistory {
	return RetainedHistory{SchemaVersion: "riido-ai-server.v1", WorkspaceID: workspace, TaskID: task, HistoryScope: "server_retained", ProjectionComplete: true, Complete: false}
}

// IDs are path segments, never display keys, URLs, queries or relative paths.
func ValidThreadReadID(s string) bool {
	if s == "" || len(s) > 512 {
		return false
	}
	for _, c := range s {
		if c == '-' || c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' {
			continue
		}
		return false
	}
	return true
}

func (r ThreadReader) read(ctx context.Context, workspace, task string) (*threadCollection, error) {
	if !ValidThreadReadID(workspace) || !ValidThreadReadID(task) {
		return nil, ErrInvalidRequest
	}
	if r.Client == nil {
		return nil, ErrUnavailable
	}
	raw, status, err := r.Client.ReadTaskThreads(ctx, workspace, task)
	if err != nil {
		return nil, err
	}
	switch status {
	case http.StatusOK:
	case http.StatusUnauthorized:
		return nil, ErrAuthentication
	case http.StatusForbidden:
		return nil, ErrPermission
	case http.StatusNotFound:
		return nil, ErrNotFound
	case http.StatusTooManyRequests:
		return nil, ErrRateLimited
	default:
		return nil, ErrUnavailable
	}
	if len(raw) > 16<<20 || !utf8.Valid(raw) {
		return nil, ErrInvalidResponse
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	if !uniqueThreadJSON(dec, 0) || dec.Decode(new(any)) != io.EOF {
		return nil, ErrInvalidResponse
	}
	var collection threadCollection
	if !threadJSONShape(raw, reflect.TypeFor[threadCollection]()) {
		return nil, ErrInvalidResponse
	}
	dec = json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if dec.Decode(&collection) != nil || collection.SchemaVersion != "riido-ai-server.v1" || collection.TaskID != task || collection.Threads == nil {
		return nil, ErrInvalidResponse
	}
	if !validThreadProjection(raw, collection, workspace, task) {
		return nil, ErrInvalidResponse
	}
	return &collection, nil
}

// Go's JSON decoder otherwise accepts duplicate/case-aliased keys. Inspect
// the original bytes before typed decoding; bound nesting as well as bytes.
func uniqueThreadJSON(dec *json.Decoder, depth int) bool {
	if depth > 100 {
		return false
	}
	token, err := dec.Token()
	if err != nil {
		return false
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return true
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for dec.More() {
			keyToken, keyErr := dec.Token()
			key, valid := keyToken.(string)
			if keyErr != nil || !valid || seen[key] {
				return false
			}
			seen[key] = true
			if !uniqueThreadJSON(dec, depth+1) {
				return false
			}
		}
		end, endErr := dec.Token()
		return endErr == nil && end == json.Delim('}')
	case '[':
		for dec.More() {
			if !uniqueThreadJSON(dec, depth+1) {
				return false
			}
		}
		end, endErr := dec.Token()
		return endErr == nil && end == json.Delim(']')
	default:
		return false
	}
}

func hasThreadFields(raw []byte, keys ...string) bool {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil {
		return false
	}
	for _, key := range keys {
		if len(fields[key]) == 0 {
			return false
		}
	}
	return true
}

func validThreadProjection(raw []byte, c threadCollection, workspace, task string) bool {
	if !hasThreadFields(raw, "schema_version", "task_id", "assigned_agent_ids", "threads") {
		return false
	}
	var shape struct {
		Threads []json.RawMessage `json:"threads"`
	}
	if json.Unmarshal(raw, &shape) != nil {
		return false
	}
	seen := map[string]bool{}
	for i, t := range c.Threads {
		if !hasThreadFields(shape.Threads[i], "thread_id", "conversation_id", "task_id", "agent_id", "run_id", "work_status", "assignment_state", "messages") || t.ThreadID == "" || seen[t.ThreadID] || t.ConversationID == "" || t.TaskID != task || t.AgentID == "" {
			return false
		}
		seen[t.ThreadID] = true
		if t.AgentSnapshotID != "" {
			snapshot := c.AgentSnapshots[t.AgentSnapshotID]
			if snapshot == nil || snapshot.AgentID != t.AgentID || (snapshot.WorkspaceID != "" && snapshot.WorkspaceID != workspace) {
				return false
			}
		}
		messageIDs := map[string]bool{}
		for _, m := range t.Messages {
			if m.MessageID == "" || messageIDs[m.MessageID] || (m.Role != "user" && m.Role != "agent" && m.Role != "progress") {
				return false
			}
			messageIDs[m.MessageID] = true
		}
		if len(t.ExecutionContext) > 0 && string(t.ExecutionContext) != "null" && bytes.TrimSpace(t.ExecutionContext)[0] != '{' {
			return false
		}
	}
	for _, snapshot := range c.AgentSnapshots {
		if snapshot == nil || snapshot.AgentID == "" || (snapshot.WorkspaceID != "" && snapshot.WorkspaceID != workspace) {
			return false
		}
	}
	return true
}
