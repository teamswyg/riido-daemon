package userauth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fakeStore struct {
	lock       chan struct{}
	credential Credential
	present    bool
	loads      int
}

func newFakeStore(credential Credential) *fakeStore {
	return &fakeStore{lock: make(chan struct{}, 1), credential: credential, present: credential.AccessToken != ""}
}

func (s *fakeStore) Lock(ctx context.Context) (func(), error) {
	select {
	case s.lock <- struct{}{}:
		return func() { <-s.lock }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *fakeStore) Load(context.Context) (Credential, error) {
	s.loads++
	if !s.present {
		return Credential{}, ErrNotAuthenticated
	}
	return s.credential, nil
}

func (s *fakeStore) Save(_ context.Context, c Credential) error {
	s.credential = c
	s.present = true
	return nil
}

func (s *fakeStore) Delete(context.Context) error {
	s.credential = Credential{}
	s.present = false
	return nil
}

func validCredential() Credential {
	return Credential{SessionID: "fixture-session", AccessToken: "old-access", RefreshToken: "old-refresh", ExpiresAt: time.Now().Add(time.Hour), TokenType: "Bearer"}
}

func TestManagersShareOneRefreshTransaction(t *testing.T) {
	var refreshes atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		refreshes.Add(1)
		fmt.Fprint(w, `{"access_token":"new-access","refresh_token":"new-refresh","token_type":"Bearer","expires_in":3600,"refresh_recovery":"v1"}`)
	}))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	credential := validCredential()
	credential.ExpiresAt = time.Now().Add(-time.Minute)
	store := newFakeStore(credential)
	managers := []*Manager{NewManager(client, store), NewManager(client, store)}
	var group sync.WaitGroup
	for i := range 16 {
		group.Add(1)
		go func(i int) {
			defer group.Done()
			token, err := managers[i%2].AccessToken(context.Background())
			if err != nil || token != "new-access" {
				t.Errorf("failed shared refresh: %v", err)
			}
		}(i)
	}
	group.Wait()
	if store.credential.SessionID != credential.SessionID {
		t.Fatal("refresh changed login session identity")
	}
	if refreshes.Load() != 1 {
		t.Fatalf("refresh calls=%d", refreshes.Load())
	}
}

func TestLogoutWinsInFlightRefresh(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	var revocations atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/revoke" {
			revocations.Add(1)
			w.WriteHeader(http.StatusOK)
			return
		}
		close(entered)
		<-release
		fmt.Fprint(w, `{"access_token":"new-access","refresh_token":"new-refresh","token_type":"Bearer","expires_in":3600,"refresh_recovery":"v1"}`)
	}))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	credential := validCredential()
	credential.ExpiresAt = time.Now().Add(-time.Minute)
	store := newFakeStore(credential)
	manager := NewManager(client, store)
	readResult := make(chan error, 1)
	go func() { _, err := manager.AccessToken(context.Background()); readResult <- err }()
	<-entered
	logoutResult := make(chan error, 1)
	go func() { logoutResult <- manager.Logout(context.Background()) }()
	deadline := time.Now().Add(time.Second)
	for manager.generation.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if manager.generation.Load() == 0 {
		t.Fatal("logout did not advance authentication generation")
	}
	close(release)
	if err := <-readResult; !errors.Is(err, ErrLoggedOut) {
		t.Fatalf("refresh after logout=%v", err)
	}
	if err := <-logoutResult; err != nil {
		t.Fatal(err)
	}
	if store.present || revocations.Load() == 0 {
		t.Fatal("logged out credential was restored")
	}
}

func TestAPIRetries401OnceButNot403(t *testing.T) {
	for _, status := range []int{401, 403} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			var calls, refreshes atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/oauth/token" {
					refreshes.Add(1)
					fmt.Fprint(w, `{"access_token":"new-access","refresh_token":"new-refresh","token_type":"Bearer","expires_in":3600,"refresh_recovery":"v1"}`)
					return
				}
				n := calls.Add(1)
				if n == 1 {
					w.WriteHeader(status)
					return
				}
				if r.Header.Get("Authorization") != "Bearer new-access" {
					t.Error("wrong refreshed authorization")
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()
			client, _ := NewClient(configFor(server))
			manager := NewManager(client, newFakeStore(validCredential()))
			request, _ := http.NewRequest(http.MethodGet, server.URL+"/work", nil)
			response, err := manager.Do(context.Background(), request)
			if err != nil {
				t.Fatal(err)
			}
			response.Body.Close()
			if status == 401 && (calls.Load() != 2 || refreshes.Load() != 1 || response.StatusCode != http.StatusOK) {
				t.Fatal("401 did not retry exactly once")
			}
			if status == 403 && (calls.Load() != 1 || refreshes.Load() != 0 || response.StatusCode != http.StatusForbidden) {
				t.Fatal("403 caused reauthentication")
			}
		})
	}
}

func TestAPIRejectsOtherOriginBeforeCredentialAccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	store := newFakeStore(validCredential())
	manager := NewManager(client, store)
	request, _ := http.NewRequest(http.MethodGet, "https://example.com/work", nil)
	response, err := manager.Do(context.Background(), request)
	if response != nil {
		response.Body.Close()
	}
	if err == nil || store.loads != 0 {
		t.Fatal("cross-origin request accessed credentials")
	}
}

func TestRefreshRevokedDeletesCredentialButNetworkDoesNot(t *testing.T) {
	for _, status := range []int{400, 503} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
				if status == 400 {
					fmt.Fprint(w, `{"error":"invalid_grant","error_description":"secret"}`)
				}
			}))
			defer server.Close()
			client, _ := NewClient(configFor(server))
			credential := validCredential()
			credential.ExpiresAt = time.Now()
			store := newFakeStore(credential)
			manager := NewManager(client, store)
			_, err := manager.AccessToken(context.Background())
			if err == nil {
				t.Fatal("failed refresh returned success")
			}
			if status == 400 && (!errors.Is(err, ErrReauthRequired) || store.present) {
				t.Fatal("revoked authentication retained")
			}
			if status == 503 && !store.present {
				t.Fatal("temporary failure erased authentication")
			}
		})
	}
}
