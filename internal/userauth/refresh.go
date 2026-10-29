package userauth

import (
	"context"
	"errors"
	"time"
)

func (m *Manager) AccessToken(ctx context.Context) (string, error) {
	credential, err := m.accessCredential(ctx, "", "")
	return credential.AccessToken, err
}

// The store lock spans reload, network rotation and save. This is required
// across CLI processes, not merely goroutines or Manager instances.
func (m *Manager) accessCredential(ctx context.Context, rejected, expectedSession string) (Credential, error) {
	generation := m.generation.Load()
	unlock, err := m.lock(ctx)
	if err != nil {
		return Credential{}, err
	}
	defer unlock()
	if m.generation.Load() != generation {
		return Credential{}, ErrLoggedOut
	}
	credential, err := m.store.Load(ctx)
	if err != nil {
		return Credential{}, storageError(err)
	}
	if m.generation.Load() != generation {
		return Credential{}, ErrLoggedOut
	}
	if expectedSession != "" && credential.SessionID != expectedSession {
		return Credential{}, ErrLoggedOut
	}
	if credential.SessionID == "" {
		return Credential{}, ErrReauthRequired
	}
	if credential.AccessToken == "" {
		return Credential{}, ErrNotAuthenticated
	}
	mustRefresh := !credential.ExpiresAt.After(time.Now().Add(15 * time.Second))
	if rejected != "" && rejected == credential.AccessToken {
		mustRefresh = true
	}
	if !mustRefresh {
		return credential, nil
	}
	fresh, err := m.client.Refresh(ctx, credential.RefreshToken)
	if err != nil {
		if errors.Is(err, ErrReauthRequired) {
			if deleteErr := m.store.Delete(ctx); deleteErr != nil {
				return Credential{}, storageError(deleteErr)
			}
		}
		return Credential{}, err
	}
	if m.generation.Load() != generation {
		// The old session cannot be restored after a local logout or account switch.
		_ = m.client.Revoke(ctx, revocationToken(fresh))
		return Credential{}, ErrLoggedOut
	}
	fresh.SessionID = credential.SessionID
	if fresh.Scope == "" {
		fresh.Scope = credential.Scope
	}
	if err := m.store.Save(ctx, fresh); err != nil {
		_ = m.client.Revoke(ctx, revocationToken(fresh))
		return Credential{}, storageError(err)
	}
	return fresh, nil
}
