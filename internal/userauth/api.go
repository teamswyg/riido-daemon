package userauth

import (
	"context"
	"io"
	"net/http"
	"strings"
)

// Do adds a user token only to the configured API origin. Only GET/HEAD are
// retried after 401; replay safety of writes or GraphQL POST is not inferred.
func (m *Manager) Do(ctx context.Context, request *http.Request) (*http.Response, error) {
	return m.do(ctx, request, false)
}

func (m *Manager) do(ctx context.Context, request *http.Request, readPOST bool) (*http.Response, error) {
	if m.client == nil || request == nil || !m.client.permits(request.URL) || request.Host != "" && !strings.EqualFold(request.Host, request.URL.Host) {
		return nil, failure("UNSAFE_REQUEST")
	}
	credential, err := m.accessCredential(ctx, "", "")
	if err != nil {
		return nil, err
	}
	response, err := m.send(ctx, request, credential.AccessToken, false)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusUnauthorized || (request.Method != http.MethodGet && request.Method != http.MethodHead && !readPOST) {
		return response, nil
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
	_ = response.Body.Close()
	credential, err = m.accessCredential(ctx, credential.AccessToken, credential.SessionID)
	if err != nil {
		return nil, err
	}
	return m.send(ctx, request, credential.AccessToken, true)
}

func (m *Manager) send(ctx context.Context, request *http.Request, token string, retry bool) (*http.Response, error) {
	outbound := request.Clone(ctx)
	if retry && request.Body != nil && request.Body != http.NoBody {
		if request.GetBody == nil {
			return nil, failure("UNSAFE_REQUEST")
		}
		body, err := request.GetBody()
		if err != nil {
			return nil, failure("UNSAFE_REQUEST")
		}
		outbound.Body = body
	}
	outbound.Header.Set("Authorization", "Bearer "+token)
	response, err := m.client.http.Do(outbound)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, &Error{Code: "HTTP_ERROR", Retryable: true}
	}
	return response, nil
}
