package userauth

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

const deviceLinkLimit = 4 << 10

func (c *Client) deviceLinkPOST(ctx context.Context, path string, body map[string]string, bearer string) ([]byte, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}
	raw, err := json.Marshal(body)
	if err != nil || len(raw) > deviceLinkLimit {
		return nil, 0, linkFailure("LINK_INVALID_CONFIG")
	}
	callCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(callCtx, http.MethodPost, c.config.Resource+path, bytes.NewReader(raw))
	if err != nil {
		return nil, 0, linkFailure("LINK_INVALID_CONFIG")
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	if bearer != "" {
		request.Header.Set("Authorization", "Bearer "+bearer)
	}
	client := *c.http
	client.Jar = nil
	client.Timeout = 5 * time.Second
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	response, err := client.Do(request)
	if err != nil {
		if callCtx.Err() != nil {
			return nil, 0, callCtx.Err()
		}
		return nil, 0, &DeviceLinkError{Code: "LINK_HTTP_ERROR", Retryable: true}
	}
	defer response.Body.Close()
	raw, err = io.ReadAll(io.LimitReader(response.Body, deviceLinkLimit+1))
	if callCtx.Err() != nil {
		return nil, 0, callCtx.Err()
	}
	if err != nil {
		return nil, response.StatusCode, &DeviceLinkError{Code: "LINK_HTTP_ERROR", StatusCode: response.StatusCode, Retryable: true}
	}
	if len(raw) > deviceLinkLimit {
		return nil, response.StatusCode, linkFailure("LINK_INVALID_RESPONSE")
	}
	return raw, response.StatusCode, nil
}

func deviceLinkHTTPError(status int, raw []byte, recovery bool) error {
	code := "LINK_REQUEST_REJECTED"
	if recovery {
		code = "LINK_RECOVERY_REJECTED"
	}
	retry := status == 408 || status == 429 || status >= 500
	if retry {
		code = "LINK_HTTP_ERROR"
	}
	if status == 429 {
		code = "LINK_RATE_LIMITED"
	}
	// Provider descriptions are ignored. No link error maps to REAUTH_REQUIRED.
	if fields, err := deviceLinkObject(raw, "error", "error_description"); err == nil {
		var serverCode string
		if json.Unmarshal(fields["error"], &serverCode) == nil && serverCode == "installation_policy_unknown" {
			code = "LINK_POLICY_UNKNOWN"
		}
	}
	return &DeviceLinkError{Code: code, StatusCode: status, Retryable: retry}
}
