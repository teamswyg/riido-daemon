package userauth

import (
	"context"
	"errors"
	"sync/atomic"
)

type Manager struct {
	client     *Client
	store      CredentialStore
	generation atomic.Uint64
}

func NewManager(client *Client, store CredentialStore) *Manager {
	return &Manager{client: client, store: store}
}

func (m *Manager) Logout(ctx context.Context) error {
	m.generation.Add(1)
	unlock, err := m.lock(ctx)
	if err != nil {
		return err
	}
	defer unlock()
	credential, err := m.store.Load(ctx)
	if errors.Is(err, ErrNotAuthenticated) {
		return nil
	}
	if err != nil {
		return storageError(err)
	}
	if err := m.store.Delete(ctx); err != nil {
		return storageError(err)
	}
	return m.client.Revoke(ctx, revocationToken(credential))
}

func (m *Manager) lock(ctx context.Context) (func(), error) {
	if m.client == nil || m.store == nil {
		return nil, failure("STORAGE_ERROR")
	}
	unlock, err := m.store.Lock(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, storageError(err)
	}
	if unlock == nil {
		return nil, failure("STORAGE_ERROR")
	}
	return unlock, nil
}

func revocationToken(credential Credential) string {
	if credential.RefreshToken != "" {
		return credential.RefreshToken
	}
	return credential.AccessToken
}
