package userauth

import (
	"context"
	"crypto/rand"
	"errors"
	"strings"
	"time"
)

// Login commits the new session before retiring the old grant. A partial
// cleanup error means the new session remains valid and must not be revoked.
func (m *Manager) Login(ctx context.Context, credential Credential) error {
	if credential.AccessToken == "" || !strings.EqualFold(credential.TokenType, "Bearer") || !credential.ExpiresAt.After(time.Now()) {
		return failure("INVALID_RESPONSE")
	}
	generation := m.generation.Add(1)
	// A deliberate login creates a new identity even if a provider reuses tokens.
	credential.SessionID = rand.Text()
	credential.RefreshRecoverySecret = ""
	unlock, err := m.lock(ctx)
	if err != nil {
		return err
	}
	defer unlock()
	previous, loadErr := m.store.Load(ctx)
	if loadErr != nil && !errors.Is(loadErr, ErrNotAuthenticated) {
		return storageError(loadErr)
	}
	if m.generation.Load() != generation {
		return ErrLoggedOut
	}
	if err := m.store.Save(ctx, credential); err != nil {
		return storageError(err)
	}
	if loadErr == nil && !sharesToken(previous, credential) {
		if err := m.client.Revoke(ctx, revocationToken(previous)); err != nil {
			return ErrPreviousGrantRevocationFailed
		}
	}
	return nil
}

func sharesToken(previous, next Credential) bool {
	return previous.AccessToken != "" && previous.AccessToken == next.AccessToken || previous.RefreshToken != "" && previous.RefreshToken == next.RefreshToken
}
