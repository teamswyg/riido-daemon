package saasplane

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"time"
)

// AssignmentContextRequest mirrors the control-plane
// POST /v1/agents/{agent_id}/assignment-context body.
type AssignmentContextRequest struct {
	AssignmentID string `json:"assignment_id"`
	LeaseToken   string `json:"lease_token"`
	View         string `json:"view"`
	Target       string `json:"target,omitempty"`
	Cursor       string `json:"cursor,omitempty"`
	RequestID    string `json:"request_id,omitempty"`
}

// AssignmentContextSource is one source the control-plane recorded as read.
type AssignmentContextSource struct {
	Kind    string `json:"kind"`
	ID      string `json:"id,omitempty"`
	Part    string `json:"part"`
	Version string `json:"version,omitempty"`
}

// AssignmentContextResponse is the snapshot view returned to the lease holder.
type AssignmentContextResponse struct {
	AssignmentID string                    `json:"assignment_id"`
	TaskID       string                    `json:"task_id"`
	View         string                    `json:"view"`
	CapturedAt   time.Time                 `json:"captured_at"`
	Truncated    bool                      `json:"truncated,omitempty"`
	Data         json.RawMessage           `json:"data"`
	Sources      []AssignmentContextSource `json:"sources"`
}

// FetchAssignmentContext reads the assignment's context snapshot with the
// device credential. It is sent once: the server records every answer as a
// read, so a blind retry would record the read twice.
func (p *Plane) FetchAssignmentContext(ctx context.Context, agentID string, in AssignmentContextRequest) (AssignmentContextResponse, error) {
	var out AssignmentContextResponse
	err := p.postJSON(ctx, "/v1/agents/"+url.PathEscape(agentID)+"/assignment-context", in, &out)
	return out, err
}

// HTTPStatus reports the status code and server error message of a non-2xx
// control-plane answer.
func HTTPStatus(err error) (int, string, bool) {
	var statusErr httpStatusError
	if !errors.As(err, &statusErr) {
		return 0, "", false
	}
	var body struct {
		Error string `json:"error"`
	}
	if json.Unmarshal([]byte(statusErr.Body), &body) != nil || body.Error == "" {
		body.Error = statusErr.Body
	}
	return statusErr.StatusCode, body.Error, true
}
