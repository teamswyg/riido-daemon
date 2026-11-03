package userauth

import (
	"context"
	"time"
)

// cachedAccess returns the cached access grant when it is still valid for this
// login (RIID-6014). Any doubt is a miss: the caller then reads the protected
// credential, which asks for user presence and refreshes when needed.
func (m *Manager) cachedAccess(ctx context.Context, expectedSession string, generation uint64) (Credential, bool) {
	cache, ok := m.store.(AccessCache)
	if !ok {
		return Credential{}, false
	}
	grant, err := cache.LoadAccess(ctx)
	if err != nil || grant.SessionID == "" || grant.AccessToken == "" || !grant.ExpiresAt.After(time.Now().Add(15*time.Second)) {
		return Credential{}, false
	}
	credential := grant.credential()
	if !m.acceptsBinding(credential) || expectedSession != "" && grant.SessionID != expectedSession || m.generation.Load() != generation {
		return Credential{}, false
	}
	return credential, true
}

// rememberAccess caches a usable access grant read from the protected
// credential. A failed save only means the next command asks again.
func (m *Manager) rememberAccess(ctx context.Context, credential Credential) {
	if cache, ok := m.store.(AccessCache); ok && credential.RefreshRecoverySecret == "" {
		_ = cache.SaveAccess(ctx, credential.AccessGrant())
	}
}

// cachedSession reports the session of the cached grant without a
// user-presence prompt. ok is false when the cache cannot answer.
func (m *Manager) cachedSession(ctx context.Context) (string, bool) {
	cache, ok := m.store.(AccessCache)
	if !ok {
		return "", false
	}
	grant, err := cache.LoadAccess(ctx)
	if err != nil || grant.SessionID == "" {
		return "", false
	}
	return grant.SessionID, true
}
