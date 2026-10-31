package userauth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"
)

var (
	ErrThreadReadUnsupported = errors.New("thread reads require a supported Riido API origin")
	ErrThreadReadIncomplete  = errors.New("retained thread history exceeds the response limit")
)

const threadReadBodyLimit = 16 << 20

// ReadTaskThreads is the only cross-origin user read. The original OAuth token
// stays at monolith; CP receives a freshly exchanged, task-scoped opaque ticket.
// No caller URL, response URL, cookie, redirect or retry can choose its target.
// The protected lock covers both hops and the bounded response read, so another
// process cannot replace the account between minting and consuming the ticket.
func (m *Manager) ReadTaskThreads(ctx context.Context, workspace, task string) ([]byte, int, error) {
	origin := threadReadOrigin(m.client)
	if origin == "" {
		return nil, 0, ErrThreadReadUnsupported
	}
	if !threadReadID(workspace) || !threadReadID(task) {
		return nil, 0, failure("UNSAFE_REQUEST")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	generation := m.generation.Load()
	unlock, err := m.lock(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer unlock()
	credential, err := m.accessCredentialLocked(ctx, "", "", generation)
	if err != nil {
		return nil, 0, err
	}
	body, err := json.Marshal(map[string]string{"workspace_id": workspace, "task_id": task})
	if err != nil {
		return nil, 0, failure("INVALID_RESPONSE")
	}
	raw, status, err := m.threadReadRequest(ctx, http.MethodPost, m.client.config.Resource+"/oauth/cli/thread-read-tickets", credential.AccessToken, body, deviceLinkLimit)
	if err != nil || status != http.StatusOK {
		return nil, status, err
	}
	ticket, err := decodeThreadReadTicket(raw, origin)
	if err != nil {
		return nil, 0, err
	}
	if m.generation.Load() != generation {
		return nil, 0, ErrLoggedOut
	}
	raw, status, err = m.threadReadRequest(ctx, http.MethodGet, origin+"/v3/client/workspaces/"+workspace+"/ai-agent/tasks/"+task+"/threads", ticket, nil, threadReadBodyLimit)
	if m.generation.Load() != generation {
		return nil, 0, ErrLoggedOut
	}
	return raw, status, err
}

func threadReadOrigin(c *Client) string {
	if !validDeviceLinkClient(c) || c.config.Endpoints.TokenURL != c.config.Resource+"/oauth/token" || c.config.Endpoints.RevokeURL != c.config.Resource+"/oauth/revoke" || c.config.Endpoints.AuthorizeURL != c.config.Resource+"/oauth/authorize" {
		return ""
	}
	switch c.config.Resource {
	case "https://monolith.riido.io":
		return "https://ai-api.riido.io"
	case "https://monolith.api.riido.io":
		return "https://staging.ai-api.riido.io"
	default:
		return ""
	}
}

func threadReadID(s string) bool {
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

func decodeThreadReadTicket(raw []byte, origin string) (string, error) {
	fields, err := deviceLinkObject(raw, "schema_version", "access_token", "token_type", "resource", "purpose", "expires_at")
	if err != nil {
		return "", failure("INVALID_RESPONSE")
	}
	var ticket struct {
		SchemaVersion string `json:"schema_version"`
		AccessToken   string `json:"access_token"`
		TokenType     string `json:"token_type"`
		Resource      string `json:"resource"`
		Purpose       string `json:"purpose"`
		ExpiresAt     string `json:"expires_at"`
	}
	if decodeDeviceLink(fields, &ticket) != nil || ticket.SchemaVersion != "riido-cli-thread-read.v1" || ticket.Resource != origin || ticket.Purpose != "task_threads_read" || ticket.TokenType != "Bearer" || !strings.HasPrefix(ticket.AccessToken, "riido_cli_thread_v1_") || !deviceLinkRandom(strings.TrimPrefix(ticket.AccessToken, "riido_cli_thread_v1_")) {
		return "", failure("INVALID_RESPONSE")
	}
	expires, err := time.Parse(time.RFC3339Nano, ticket.ExpiresAt)
	now := time.Now()
	// The server issues at most 60 seconds; tolerate at most 5 seconds of skew.
	if err != nil || !expires.After(now) || expires.After(now.Add(65*time.Second)) {
		return "", failure("INVALID_RESPONSE")
	}
	return ticket.AccessToken, nil
}

func (m *Manager) threadReadRequest(ctx context.Context, method, endpoint, token string, body []byte, limit int64) ([]byte, int, error) {
	request, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, 0, failure("UNSAFE_REQUEST")
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Accept", "application/json")
	if method == http.MethodPost {
		request.Header.Set("Content-Type", "application/json")
	}
	client := *m.client.http
	client.Jar = nil
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	response, err := client.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return nil, 0, ctx.Err()
		}
		return nil, 0, failure("HTTP_ERROR")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, response.StatusCode, nil
	}
	mediaType, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return nil, 0, failure("INVALID_RESPONSE")
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil {
		if ctx.Err() != nil {
			return nil, 0, ctx.Err()
		}
		return nil, 0, failure("HTTP_ERROR")
	}
	if int64(len(raw)) > limit {
		if limit == threadReadBodyLimit {
			return nil, 0, ErrThreadReadIncomplete
		}
		return nil, 0, failure("INVALID_RESPONSE")
	}
	return raw, response.StatusCode, nil
}
