package saasplane

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
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
	Since        string `json:"since,omitempty"`
	RequestID    string `json:"request_id,omitempty"`
	// MaxBytes asks a content read for consecutive pages of one part up to
	// this encoded size. Zero keeps the single-page response.
	MaxBytes     int    `json:"max_bytes,omitempty"`
	CacheEnabled bool   `json:"-"`
	KnownDigest  string `json:"-"`
}

// AssignmentContextSource is one source the control-plane recorded as read.
type AssignmentContextSource struct {
	Kind      string `json:"kind"`
	ID        string `json:"id,omitempty"`
	Part      string `json:"part"`
	Version   string `json:"version,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// AssignmentContextResponse is the snapshot (or live, for the agent
// directory) view returned to the lease holder.
type AssignmentContextResponse struct {
	ContentDigest string                    `json:"content_digest,omitempty"`
	NotModified   bool                      `json:"not_modified,omitempty"`
	AssignmentID  string                    `json:"assignment_id"`
	TaskID        string                    `json:"task_id"`
	View          string                    `json:"view"`
	CapturedAt    time.Time                 `json:"captured_at"`
	Truncated     bool                      `json:"truncated,omitempty"`
	Live          bool                      `json:"live,omitempty"`
	Data          json.RawMessage           `json:"data"`
	Sources       []AssignmentContextSource `json:"sources"`
}

// FetchAssignmentContext reads the assignment's context snapshot with the
// device credential. It is sent once: the server records every answer as a
// read, so a blind retry would record the read twice.
func (p *Plane) FetchAssignmentContext(ctx context.Context, agentID string, in AssignmentContextRequest) (AssignmentContextResponse, error) {
	var out AssignmentContextResponse
	body, err := json.Marshal(in)
	if err != nil {
		return out, err
	}
	headers := make(http.Header)
	if in.View == "changes" {
		// Opt in to deleted/unselected omissions; older servers ignore it.
		headers.Set("X-Riido-Context-Changes-Version", "2")
	}
	if in.CacheEnabled && in.View == "content" {
		headers.Set("X-Riido-Context-Cache-Version", "1")
		if in.KnownDigest != "" {
			headers.Set("X-Riido-Context-Known-Digest", in.KnownDigest)
		}
	}
	err = p.doJSONHeaders(ctx, http.MethodPost, "/v1/agents/"+url.PathEscape(agentID)+"/assignment-context", body, &out, headers)
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
