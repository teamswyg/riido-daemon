package userauth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestLoginStoresNewSessionBeforePreviousGrantRevokeFailure(t *testing.T) {
	store := newFakeStore(validCredential())
	next := validCredential()
	next.AccessToken = "new-access"
	next.RefreshToken = "new-refresh"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if store.credential.AccessToken != "new-access" {
			t.Error("new session was not saved before old revocation")
		}
		if err := r.ParseForm(); err != nil {
			t.Error(err)
			return
		}
		if r.Form.Get("token") != "old-refresh" {
			t.Error("wrong previous grant token")
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	err := NewManager(client, store).Login(context.Background(), next)
	if !errors.Is(err, ErrPreviousGrantRevocationFailed) {
		t.Fatalf("missing partial-cleanup status: %v", err)
	}
	if store.credential.AccessToken != "new-access" {
		t.Fatal("remote failure lost new authenticated session")
	}
	if store.loads == 0 {
		t.Fatal("login did not reload existing grant")
	}
}

func TestLoginDoesNotRevokeSharedToken(t *testing.T) {
	for _, shared := range []string{"refresh", "access"} {
		t.Run(shared, func(t *testing.T) {
			var revocations atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { revocations.Add(1) }))
			defer server.Close()
			old := validCredential()
			next := old
			if shared == "refresh" {
				next.AccessToken = "new-access"
			} else {
				next.RefreshToken = "new-refresh"
			}
			client, _ := NewClient(configFor(server))
			store := newFakeStore(old)
			if err := NewManager(client, store).Login(context.Background(), next); err != nil {
				t.Fatal(err)
			}
			if revocations.Load() != 0 {
				t.Fatal("revoked a token shared with the new session")
			}
		})
	}
}

type rejectingSaveStore struct{ *fakeStore }

func (*rejectingSaveStore) Save(context.Context, Credential) error {
	return errors.New("secret-store-implementation-detail")
}

func TestLoginSaveFailurePreservesOldSession(t *testing.T) {
	var revocations atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { revocations.Add(1) }))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	store := &rejectingSaveStore{newFakeStore(validCredential())}
	next := validCredential()
	next.AccessToken = "new-access"
	next.RefreshToken = "new-refresh"
	err := NewManager(client, store).Login(context.Background(), next)
	if err == nil || err.Error() == "secret-store-implementation-detail" {
		t.Fatal("save failure was not safely reported")
	}
	if store.credential.AccessToken != "old-access" || revocations.Load() != 0 {
		t.Fatal("save failure changed or revoked old session")
	}
}
