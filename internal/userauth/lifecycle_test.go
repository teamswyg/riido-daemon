package userauth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLogoutClearsLocalBeforeRemoteFailure(t *testing.T) {
	store := newFakeStore(validCredential())
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if store.present {
			t.Error("remote revoke began before local credential deletion")
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	manager := NewManager(client, store)
	if err := manager.Logout(context.Background()); err == nil {
		t.Fatal("remote revoke failure was hidden")
	}
	if store.present {
		t.Fatal("remote failure retained local credential")
	}
}

func TestCrossManagerLogoutRevokesRotatedToken(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	revoked := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/revoke" {
			if err := r.ParseForm(); err != nil {
				t.Error(err)
				return
			}
			revoked <- r.Form.Get("token")
			return
		}
		close(entered)
		<-release
		fmt.Fprint(w, `{"access_token":"rotated-access","refresh_token":"rotated-refresh","token_type":"Bearer","expires_in":3600,"refresh_recovery":"v1"}`)
	}))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	credential := validCredential()
	credential.ExpiresAt = time.Now().Add(-time.Minute)
	store := newFakeStore(credential)
	first, second := NewManager(client, store), NewManager(client, store)
	refreshed := make(chan error, 1)
	go func() { _, err := first.AccessToken(context.Background()); refreshed <- err }()
	<-entered
	loggedOut := make(chan error, 1)
	go func() { loggedOut <- second.Logout(context.Background()) }()
	close(release)
	if err := <-refreshed; err != nil {
		t.Fatal(err)
	}
	if err := <-loggedOut; err != nil {
		t.Fatal(err)
	}
	if token := <-revoked; token != "rotated-refresh" {
		t.Fatal("logout did not reload rotated credential")
	}
	if _, err := first.AccessToken(context.Background()); !errors.Is(err, ErrNotAuthenticated) {
		t.Fatal("other manager retained logged out credential")
	}
}

func TestRefreshPreservesScopeWhenOmitted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"access_token":"new-access","refresh_token":"new-refresh","token_type":"Bearer","expires_in":3600,"refresh_recovery":"v1"}`)
	}))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	credential := validCredential()
	credential.Scope = "read:workspace"
	credential.ExpiresAt = time.Now()
	store := newFakeStore(credential)
	if _, err := NewManager(client, store).AccessToken(context.Background()); err != nil {
		t.Fatal(err)
	}
	if store.credential.Scope != "read:workspace" || store.credential.RefreshToken != "new-refresh" {
		t.Fatal("refresh omitted fields erased existing grant metadata")
	}
}
