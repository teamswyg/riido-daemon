package userauth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

type delayedLoadStore struct {
	*fakeStore
	entered, release chan struct{}
	once             sync.Once
}

func (s *delayedLoadStore) Load(ctx context.Context) (Credential, error) {
	s.once.Do(func() { close(s.entered); <-s.release })
	return s.fakeStore.Load(ctx)
}

func TestLogoutDuringUserPresenceCheckRejectsLoadedToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	store := &delayedLoadStore{fakeStore: newFakeStore(validCredential()), entered: make(chan struct{}), release: make(chan struct{})}
	manager := NewManager(client, store)
	loaded := make(chan error, 1)
	go func() { _, err := manager.AccessToken(context.Background()); loaded <- err }()
	<-store.entered
	loggedOut := make(chan error, 1)
	go func() { loggedOut <- manager.Logout(context.Background()) }()
	deadline := time.Now().Add(time.Second)
	for manager.generation.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	close(store.release)
	if err := <-loaded; !errors.Is(err, ErrLoggedOut) {
		t.Errorf("token released after logout began: %v", err)
	}
	if err := <-loggedOut; err != nil {
		t.Fatal(err)
	}
}

func TestStorageCancellationPreservesCallerContext(t *testing.T) {
	for _, expected := range []error{context.Canceled, context.DeadlineExceeded} {
		if !errors.Is(storageError(expected), expected) {
			t.Fatal("credential store cancellation was masked")
		}
	}
}
