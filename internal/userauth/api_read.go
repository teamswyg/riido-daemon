package userauth

import (
	"context"
	"net/http"
)

// DoRead permits one 401 retry for a known read-only GraphQL POST. Callers must
// supply a fixed query, never a mutation or user-provided operation document.
// GetBody is mandatory so the exact request can be replayed safely.
func (m *Manager) DoRead(ctx context.Context, request *http.Request) (*http.Response, error) {
	if request == nil {
		return nil, failure("UNSAFE_REQUEST")
	}
	if request.Method == http.MethodGet || request.Method == http.MethodHead {
		return m.Do(ctx, request)
	}
	if request.Method != http.MethodPost || request.Body == nil || request.GetBody == nil {
		return nil, failure("UNSAFE_REQUEST")
	}
	return m.do(ctx, request, true)
}
