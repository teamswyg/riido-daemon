package userauth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const responseLimit = 1 << 20

func (c *Client) formRequest(ctx context.Context, endpoint string, form url.Values, result any) error {
	form.Set("client_id", c.config.ClientID)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return failure("INVALID_CONFIG")
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")
	response, err := c.http.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return &Error{Code: "HTTP_ERROR", Retryable: true}
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, responseLimit+1))
	if err != nil || len(body) > responseLimit {
		return failure("INVALID_RESPONSE")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return oauthError(response.StatusCode, body)
	}
	if result == nil {
		return nil
	}
	if err := json.Unmarshal(body, result); err != nil {
		return failure("INVALID_RESPONSE")
	}
	return nil
}

func oauthError(status int, body []byte) error {
	var payload struct {
		Code string `json:"error"`
	}
	_ = json.Unmarshal(body, &payload)
	if payload.Code == "invalid_grant" || payload.Code == "invalid_token" {
		return &Error{Code: "REAUTH_REQUIRED", StatusCode: status}
	}
	if status >= 500 || status == 429 {
		return &Error{Code: "HTTP_ERROR", StatusCode: status, Retryable: true}
	}
	return &Error{Code: "REQUEST_REJECTED", StatusCode: status}
}
