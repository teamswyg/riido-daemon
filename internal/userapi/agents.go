package userapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"unicode/utf8"
)

func (r AgentReader) List(ctx context.Context, workspace, task, cursor string) (*AgentPage, error) {
	if !ValidThreadReadID(workspace) || !ValidThreadReadID(task) || !ValidAgentCursor(cursor) {
		return nil, ErrInvalidRequest
	}
	if r.Client == nil {
		return nil, ErrUnavailable
	}
	var page AgentPage
	err := r.Client.ReadTaskAgents(ctx, workspace, task, "", cursor, func(raw []byte, status int) error {
		if err := decodeAgentResponse(raw, status, &page); err != nil {
			return err
		}
		if !validAgentPage(raw, page, cursor) {
			return ErrInvalidResponse
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &page, nil
}

func (r AgentReader) Show(ctx context.Context, workspace, task, id string) (*AgentDetail, error) {
	if !ValidThreadReadID(workspace) || !ValidThreadReadID(task) || !ValidThreadReadID(id) {
		return nil, ErrInvalidRequest
	}
	if r.Client == nil {
		return nil, ErrUnavailable
	}
	var detail AgentDetail
	err := r.Client.ReadTaskAgents(ctx, workspace, task, id, "", func(raw []byte, status int) error {
		if err := decodeAgentResponse(raw, status, &detail); err != nil {
			return err
		}
		var shape struct {
			Agent json.RawMessage `json:"agent"`
		}
		if json.Unmarshal(raw, &shape) != nil || !hasThreadFields(raw, "agent", "fetched_at", "policy_checked") || detail.FetchedAt.IsZero() || detail.PolicyChecked || detail.Agent.AgentID != id || !validPublicAgent(shape.Agent, detail.Agent) {
			return ErrInvalidResponse
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &detail, nil
}

func ValidAgentCursor(cursor string) bool {
	if cursor == "" {
		return true
	}
	n, err := strconv.Atoi(cursor)
	return err == nil && n >= 0 && strconv.Itoa(n) == cursor
}

func decodeAgentResponse(raw []byte, status int, out any) error {
	switch status {
	case http.StatusOK:
	case http.StatusUnauthorized:
		return ErrAuthentication
	case http.StatusForbidden:
		return ErrPermission
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusTooManyRequests:
		return ErrRateLimited
	default:
		return ErrUnavailable
	}
	if len(raw) > 16<<20 || !utf8.Valid(raw) {
		return ErrInvalidResponse
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	if !uniqueThreadJSON(dec, 0) || dec.Decode(new(any)) != io.EOF || !threadJSONShape(raw, reflect.TypeOf(out).Elem()) {
		return ErrInvalidResponse
	}
	if json.Unmarshal(raw, out) != nil {
		return ErrInvalidResponse
	}
	return nil
}
