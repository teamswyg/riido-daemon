package userauth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestAPIRejectsSpoofedHostBeforeCredentialAccess(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1) }))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	store := newFakeStore(validCredential())
	request, _ := http.NewRequest(http.MethodGet, server.URL+"/work", nil)
	request.Host = "other.example"
	response, err := NewManager(client, store).Do(context.Background(), request)
	if response != nil {
		response.Body.Close()
	}
	if err == nil || calls.Load() != 0 || store.loads != 0 {
		t.Fatal("spoofed API host reached credential or server")
	}
}

type lockedStore struct{ *fakeStore }

func (*lockedStore) Load(context.Context) (Credential, error) { return Credential{}, ErrStoreLocked }

func TestUserPresenceRefusalPreventsNetworkRequest(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1) }))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	store := &lockedStore{newFakeStore(validCredential())}
	request, _ := http.NewRequest(http.MethodGet, server.URL+"/work", nil)
	response, err := NewManager(client, store).Do(context.Background(), request)
	if response != nil {
		response.Body.Close()
	}
	if !errors.Is(err, ErrStoreLocked) || calls.Load() != 0 {
		t.Fatal("user-presence refusal did not stop request")
	}
}
