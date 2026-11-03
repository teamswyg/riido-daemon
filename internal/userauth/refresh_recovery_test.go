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

const recoveryResponse = `{"access_token":"recovered-access","refresh_token":"recovered-refresh","token_type":"Bearer","expires_in":600,"refresh_recovery":"v1"}`

func TestRefreshRecoveryPersistedBeforeNetworkAndReusedAfterRestart(t *testing.T) {
	credential := validCredential()
	credential.ExpiresAt = time.Now()
	store := newFakeStore(credential)
	var originalSecret string
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.ParseForm() != nil || r.Form.Get("grant_type") != refreshRecoveryGrant || r.Form.Get("refresh_token") != credential.RefreshToken || r.Form.Get("recovery_secret") != store.credential.RefreshRecoverySecret || store.credential.SessionID != credential.SessionID {
			t.Error("rotation sent before saving original session/pair")
		}
		if calls == 1 {
			originalSecret = r.Form.Get("recovery_secret")
			if len(originalSecret) != 43 {
				t.Error("missing protected recovery secret")
			}
			// Simulate commit followed by a lost response, not a returned error.
			connection, _, err := w.(http.Hijacker).Hijack()
			if err != nil {
				t.Error(err)
				return
			}
			_ = connection.Close()
			return
		}
		if r.Form.Get("recovery_secret") != originalSecret {
			t.Error("restart replaced original recovery request")
		}
		fmt.Fprint(w, recoveryResponse)
	}))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	if _, err := NewManager(client, store).AccessToken(t.Context()); err == nil || store.credential.RefreshRecoverySecret == "" {
		t.Fatal("lost response erased recovery")
	}
	// A restarted process only knows the persisted record. Pending recovery
	// must run even if the old access token has not yet expired.
	restartedStore := newFakeStore(store.credential)
	restartedStore.credential.ExpiresAt = time.Now().Add(time.Hour)
	token, err := NewManager(client, restartedStore).AccessToken(t.Context())
	if err != nil || token != "recovered-access" || calls != 2 || restartedStore.credential.RefreshRecoverySecret != "" || restartedStore.credential.SessionID != credential.SessionID {
		t.Fatal("restart did not finish original rotation", err)
	}
}

type ambiguousRefreshStore struct {
	*fakeStore
	failAt    int
	saves     int
	committed bool
}

func (s *ambiguousRefreshStore) Save(ctx context.Context, credential Credential) error {
	s.saves++
	if s.saves == s.failAt {
		if s.committed {
			_ = s.fakeStore.Save(ctx, credential)
		}
		return errors.New("synthetic storage acknowledgement loss")
	}
	return s.fakeStore.Save(ctx, credential)
}

func TestRefreshRecoveryAmbiguousSavesNeverDestroyGrant(t *testing.T) {
	for _, failAt := range []int{1, 2} {
		for _, committed := range []bool{false, true} {
			t.Run(fmt.Sprintf("save%d/committed%v", failAt, committed), func(t *testing.T) {
				credential := validCredential()
				credential.ExpiresAt = time.Now()
				store := &ambiguousRefreshStore{fakeStore: newFakeStore(credential), failAt: failAt, committed: committed}
				calls, revocations := 0, 0
				var secret string
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/oauth/revoke" {
						revocations++
						return
					}
					calls++
					if r.ParseForm() != nil {
						t.Error("invalid test form")
					}
					if secret == "" {
						secret = r.Form.Get("recovery_secret")
					} else if secret != r.Form.Get("recovery_secret") {
						t.Error("ambiguous final save replaced recovery secret")
					}
					fmt.Fprint(w, recoveryResponse)
				}))
				defer server.Close()
				client, _ := NewClient(configFor(server))
				if _, err := NewManager(client, store).AccessToken(t.Context()); err == nil {
					t.Fatal("save failure reported success")
				}
				if revocations != 0 || (failAt == 1 && calls != 0) {
					t.Fatal("ambiguous save revoked a grant or sent before durable preparation")
				}
				restarted := newFakeStore(store.credential)
				if token, err := NewManager(client, restarted).AccessToken(t.Context()); err != nil || token != "recovered-access" || restarted.credential.RefreshRecoverySecret != "" {
					t.Fatal("ambiguous save prevented recovery", err)
				}
				if revocations != 0 {
					t.Fatal("recovery revoked the active session")
				}
			})
		}
	}
}

func TestRefreshRecoveryOldServerNeverFallsBackOrDeletesPending(t *testing.T) {
	credential := validCredential()
	credential.ExpiresAt = time.Now()
	store := newFakeStore(credential)
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.ParseForm() != nil || r.Form.Get("grant_type") != refreshRecoveryGrant {
			t.Error("request fell back to legacy rotation")
		}
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":"unsupported_grant_type"}`)
	}))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	for range 2 {
		if _, err := NewManager(client, store).AccessToken(t.Context()); err == nil {
			t.Fatal("old server reported success")
		}
	}
	if calls != 2 || !store.present || store.credential.RefreshRecoverySecret == "" || store.credential.RefreshToken != credential.RefreshToken {
		t.Fatal("unsupported server destroyed pending credentials or retried implicitly")
	}
}

func TestRefreshRecoveryRequiresVersionedRotationReceipt(t *testing.T) {
	for _, body := range []string{
		`{"access_token":"next","refresh_token":"next","token_type":"Bearer","expires_in":600}`,
		`{"access_token":"next","token_type":"Bearer","expires_in":600,"refresh_recovery":"v1"}`,
		`{"access_token":"next","refresh_token":"old-refresh","token_type":"Bearer","expires_in":600,"refresh_recovery":"v1"}`,
	} {
		t.Run(fmt.Sprint(len(body)), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, body) }))
			defer server.Close()
			client, _ := NewClient(configFor(server))
			credential := validCredential()
			credential.ExpiresAt = time.Now()
			store := newFakeStore(credential)
			if _, err := NewManager(client, store).AccessToken(t.Context()); err == nil || store.credential.RefreshRecoverySecret == "" || store.credential.AccessToken != credential.AccessToken {
				t.Fatal("invalid receipt committed or erased pending")
			}
		})
	}
}

func TestRefreshRecoveryPendingDoesNotCrossLogoutOrNewLogin(t *testing.T) {
	for _, operation := range []string{"logout", "login"} {
		t.Run(operation, func(t *testing.T) {
			credential := validCredential()
			credential.RefreshRecoverySecret, _ = newRefreshRecoverySecret()
			store := newFakeStore(credential)
			revokes := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/oauth/revoke" || r.ParseForm() != nil || r.Form.Get("token") != credential.RefreshToken {
					t.Error("old pending grant was not revoked directly")
				}

				revokes++
			}))
			defer server.Close()
			client, _ := NewClient(configFor(server))
			manager := NewManager(client, store)
			if operation == "logout" {
				if err := manager.Logout(t.Context()); err != nil || store.present {
					t.Fatal("logout retained pending", err)
				}
			} else {
				next := credential
				next.AccessToken, next.RefreshToken = "second-access", "second-refresh"
				if err := manager.Login(t.Context(), next); err != nil || store.credential.SessionID == credential.SessionID || store.credential.RefreshRecoverySecret != "" {
					t.Fatal("new account inherited old pending", err)
				}
			}
			if revokes != 1 {
				t.Fatal("original pending family was not retired")
			}
		})
	}
}

func TestRefreshRecoveryMissingTokenStillRequiresNewLogin(t *testing.T) {
	credential := validCredential()
	credential.ExpiresAt = time.Now()
	credential.RefreshToken = ""
	store := newFakeStore(credential)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("missing refresh reached network")
	}))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	if _, err := NewManager(client, store).AccessToken(t.Context()); !errors.Is(err, ErrReauthRequired) || store.present {
		t.Fatal("missing refresh no longer clears invalid credential", err)
	}
}
