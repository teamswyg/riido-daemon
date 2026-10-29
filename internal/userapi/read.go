package userapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

const (
	whoamiQuery = `query RiidoCLIWhoami { getMyAccountInfo { id email username loginType lastViewedWorkspaceId } }`
	workQuery   = `query RiidoCLIWork($componentId:String!,$teamId:String!){getComponent(componentId:$componentId,teamId:$teamId,getComponentDto:{getDocument:"true"}){id componentType title keyNumber description createdAt updatedAt document{id HTMLContent tiptapDocumentId} startDate endDate estimate totalPriorityScore isDone isArchived status{id statusName statusType} componentMembers{id memberOrder teamMember{id status workspaceProfile{nickname status}}} componentLabels{id labelOrder label{id labelName}}}}`
)

func (r Reader) Whoami(ctx context.Context) (Identity, error) {
	var data struct {
		Account *Identity `json:"getMyAccountInfo"`
	}
	if err := r.read(ctx, whoamiQuery, nil, &data); err != nil {
		return Identity{}, err
	}
	if data.Account == nil || data.Account.ID == "" {
		return Identity{}, ErrInvalidResponse
	}
	return *data.Account, nil
}

func (r Reader) Work(ctx context.Context, id, teamID string) (Work, error) {
	if id == "" || teamID == "" {
		return Work{}, ErrInvalidResponse
	}
	var data struct {
		Component *Work `json:"getComponent"`
	}
	if err := r.read(ctx, workQuery, map[string]string{"componentId": id, "teamId": teamID}, &data); err != nil {
		return Work{}, err
	}
	if data.Component == nil || data.Component.ID != id {
		return Work{}, ErrInvalidResponse
	}
	if data.Component.ComponentMembers == nil || data.Component.ComponentLabels == nil {
		return Work{}, ErrInvalidResponse
	}
	return *data.Component, nil
}

func (r Reader) read(ctx context.Context, query string, variables, data any) error {
	if r.Client == nil {
		return ErrUnavailable
	}
	body, err := json.Marshal(struct {
		Query     string `json:"query"`
		Variables any    `json:"variables,omitempty"`
	}{query, variables})
	if err != nil {
		return ErrInvalidResponse
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.Endpoint, bytes.NewReader(body))
	if err != nil {
		return ErrInvalidResponse
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := r.Client.DoRead(ctx, req)
	if err != nil {
		return err
	}
	if res == nil || res.Body == nil {
		return ErrInvalidResponse
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case http.StatusUnauthorized:
		return ErrAuthentication
	case http.StatusForbidden:
		return ErrPermission
	case http.StatusTooManyRequests:
		return ErrRateLimited
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return ErrUnavailable
	}
	raw, err := io.ReadAll(io.LimitReader(res.Body, maxBodyBytes+1))
	if err != nil || len(raw) > maxBodyBytes {
		return ErrInvalidResponse
	}
	var envelope struct {
		Data   json.RawMessage   `json:"data"`
		Errors []json.RawMessage `json:"errors"`
	}
	if json.Unmarshal(raw, &envelope) != nil {
		return ErrInvalidResponse
	}
	if len(envelope.Errors) > 0 {
		return classifyGraphQLErrors(envelope.Errors)
	}
	if len(envelope.Data) == 0 || bytes.Equal(envelope.Data, []byte("null")) {
		return ErrInvalidResponse
	}
	if json.Unmarshal(envelope.Data, data) != nil {
		return ErrInvalidResponse
	}
	return nil
}
