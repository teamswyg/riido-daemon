package userauth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// cachingStore mirrors the protected store contract: Save refreshes the access
// cache and Delete clears it first.
type cachingStore struct {
	*fakeStore
	grant     *AccessGrant
	loadErr   error
	saveCalls int
}

func (s *cachingStore) Save(ctx context.Context, c Credential) error {
	if err := s.fakeStore.Save(ctx, c); err != nil {
		return err
	}
	grant := c.AccessGrant()
	s.grant = &grant
	return nil
}

func (s *cachingStore) Delete(ctx context.Context) error {
	s.grant = nil
	return s.fakeStore.Delete(ctx)
}

func (s *cachingStore) LoadAccess(context.Context) (AccessGrant, error) {
	if s.loadErr != nil {
		return AccessGrant{}, s.loadErr
	}
	if s.grant == nil {
		return AccessGrant{}, ErrNotAuthenticated
	}
	return *s.grant, nil
}

func (s *cachingStore) SaveAccess(_ context.Context, grant AccessGrant) error {
	s.saveCalls++
	s.grant = &grant
	return nil
}

func cacheServer(t *testing.T, refreshes *atomic.Int32, rejectOld bool) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			refreshes.Add(1)
			fmt.Fprint(w, `{"access_token":"new-access","refresh_token":"new-refresh","token_type":"Bearer","expires_in":3600,"refresh_recovery":"v1"}`)
			return
		}
		if rejectOld && r.Header.Get("Authorization") == "Bearer old-access" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		fmt.Fprint(w, `{}`)
	}))
	t.Cleanup(server.Close)
	return server
}

// RIID-6014: a valid cached access grant serves reads without loading the
// user-presence protected credential; the first read fills the cache.
func TestReadSessionUsesAccessCacheWithoutProtectedLoad(t *testing.T) {
	var refreshes atomic.Int32
	server := cacheServer(t, &refreshes, false)
	client, _ := NewClient(configFor(server))
	store := &cachingStore{fakeStore: newFakeStore(validCredential())}
	manager := NewManager(client, store)
	read := func() error {
		return manager.WithReadSession(context.Background(), func(reader SessionReadClient) error {
			return readSessionRequest(t, reader, server.URL+"/graphql")
		})
	}
	if err := read(); err != nil || store.loads != 1 || store.saveCalls != 1 {
		t.Fatalf("first read must fill the cache: err=%v loads=%d saves=%d", err, store.loads, store.saveCalls)
	}
	for range 3 {
		if err := read(); err != nil {
			t.Fatal(err)
		}
	}
	if store.loads != 1 || refreshes.Load() != 0 {
		t.Fatalf("cached reads loaded the protected credential: loads=%d refreshes=%d", store.loads, refreshes.Load())
	}
}

func TestAccessCacheMissesFallBackToProtectedCredential(t *testing.T) {
	expired := validCredential().AccessGrant()
	expired.ExpiresAt = time.Now().Add(5 * time.Second)
	otherSession := validCredential().AccessGrant()
	otherSession.SessionID = "other-session"
	for name, setup := range map[string]func(*cachingStore){
		"expiring":    func(s *cachingStore) { s.grant = &expired },
		"unavailable": func(s *cachingStore) { s.loadErr = errors.New("helper without cache") },
		"empty":       func(s *cachingStore) { s.grant = &AccessGrant{} },
	} {
		t.Run(name, func(t *testing.T) {
			var refreshes atomic.Int32
			server := cacheServer(t, &refreshes, false)
			client, _ := NewClient(configFor(server))
			store := &cachingStore{fakeStore: newFakeStore(validCredential())}
			setup(store)
			manager := NewManager(client, store)
			token, err := manager.AccessToken(context.Background())
			if err != nil || token != "old-access" || store.loads != 1 {
				t.Fatalf("token=%q err=%v loads=%d", token, err, store.loads)
			}
		})
	}
	t.Run("expected session mismatch", func(t *testing.T) {
		var refreshes atomic.Int32
		server := cacheServer(t, &refreshes, false)
		client, _ := NewClient(configFor(server))
		store := &cachingStore{fakeStore: newFakeStore(validCredential()), grant: &otherSession}
		manager := NewManager(client, store)
		if _, ok := manager.cachedAccess(context.Background(), "fixture-session", manager.generation.Load()); ok {
			t.Fatal("cache from another login was accepted")
		}
	})
}

// A rejected cached token goes through the protected credential and refresh,
// and the refreshed grant replaces the cache.
func TestRejectedCachedTokenRefreshesThroughProtectedCredential(t *testing.T) {
	var refreshes atomic.Int32
	server := cacheServer(t, &refreshes, true)
	client, _ := NewClient(configFor(server))
	store := &cachingStore{fakeStore: newFakeStore(validCredential())}
	grant := validCredential().AccessGrant()
	store.grant = &grant
	manager := NewManager(client, store)
	err := manager.WithReadSession(context.Background(), func(reader SessionReadClient) error {
		return readSessionRequest(t, reader, server.URL+"/graphql")
	})
	if err != nil || refreshes.Load() != 1 || store.loads == 0 || store.grant == nil || store.grant.AccessToken != "new-access" {
		t.Fatalf("err=%v refreshes=%d loads=%d", err, refreshes.Load(), store.loads)
	}
}

// Logout clears the cache with the protected credential, so a later read must
// not reuse the cached token.
func TestLogoutClearsAccessCache(t *testing.T) {
	var refreshes atomic.Int32
	server := cacheServer(t, &refreshes, false)
	client, _ := NewClient(configFor(server))
	store := &cachingStore{fakeStore: newFakeStore(validCredential())}
	grant := validCredential().AccessGrant()
	store.grant = &grant
	manager := NewManager(client, store)
	if err := manager.Logout(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.AccessToken(context.Background()); !errors.Is(err, ErrNotAuthenticated) {
		t.Fatalf("read after logout: %v", err)
	}
}
