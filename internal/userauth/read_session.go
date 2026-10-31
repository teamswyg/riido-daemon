package userauth

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
)

// SessionReadClient is valid only inside WithReadSession. Callers supply fixed
// read-only operations and consume/close each response before returning.
type SessionReadClient interface {
	DoRead(context.Context, *http.Request) (*http.Response, error)
}

// WithReadSession binds a compound same-origin read to one protected login.
// The cross-process store lock covers all network calls and response parsing.
// Callers must discard their result whenever this method returns an error.
func (m *Manager) WithReadSession(ctx context.Context, read func(SessionReadClient) error) error {
	if read == nil {
		return failure("UNSAFE_REQUEST")
	}
	generation := m.generation.Load()
	unlock, err := m.lock(ctx)
	if err != nil {
		return err
	}
	defer unlock()
	credential, err := m.accessCredentialLocked(ctx, "", "", generation)
	if err != nil {
		return err
	}
	session := &sessionReadClient{manager: m, generation: generation, sessionID: credential.SessionID, active: true}
	defer session.close()
	if err := read(session); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.generation.Load() != generation {
		return ErrLoggedOut
	}
	current, err := m.store.Load(ctx)
	if err != nil {
		return storageError(err)
	}
	if current.SessionID != session.sessionID || m.generation.Load() != generation {
		return ErrLoggedOut
	}
	return nil
}

type sessionReadClient struct {
	mu         sync.Mutex
	manager    *Manager
	sessionID  string
	generation uint64
	active     bool
}

func (s *sessionReadClient) close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = false
}

func (s *sessionReadClient) DoRead(ctx context.Context, request *http.Request) (*http.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.manager
	if !s.active || m.generation.Load() != s.generation {
		return nil, ErrLoggedOut
	}
	if request == nil || !m.client.permits(request.URL) || request.Host != "" && !strings.EqualFold(request.Host, request.URL.Host) {
		return nil, failure("UNSAFE_REQUEST")
	}
	if request.Method != http.MethodGet && request.Method != http.MethodHead && (request.Method != http.MethodPost || request.Body == nil || request.GetBody == nil) {
		return nil, failure("UNSAFE_REQUEST")
	}
	credential, err := m.accessCredentialLocked(ctx, "", s.sessionID, s.generation)
	if err != nil {
		return nil, err
	}
	response, err := s.send(ctx, request, credential.AccessToken, false)
	if err != nil || response.StatusCode != http.StatusUnauthorized {
		return response, err
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
	_ = response.Body.Close()
	credential, err = m.accessCredentialLocked(ctx, credential.AccessToken, s.sessionID, s.generation)
	if err != nil {
		return nil, err
	}
	return s.send(ctx, request, credential.AccessToken, true)
}

func (s *sessionReadClient) send(ctx context.Context, request *http.Request, token string, retry bool) (*http.Response, error) {
	response, err := s.manager.send(ctx, request, token, retry)
	if s.manager.generation.Load() != s.generation {
		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}
		return nil, ErrLoggedOut
	}
	return response, err
}
