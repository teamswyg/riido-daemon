package standalonecli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

type testStore struct {
	mu             sync.Mutex
	credential     userauth.Credential
	present        bool
	statusErr      error
	saves          int
	confirmErr     error
	confirms       int
	confirmHook    func()
	order          []string
	legacyConfirms int
	signErr        error
	signs          int
}

func (s *testStore) Lock(context.Context) (func(), error) { s.mu.Lock(); return s.mu.Unlock, nil }
func (s *testStore) Load(context.Context) (userauth.Credential, error) {
	if !s.present {
		return userauth.Credential{}, userauth.ErrNotAuthenticated
	}
	return s.credential, nil
}

func (s *testStore) Save(_ context.Context, v userauth.Credential) error {
	s.order = append(s.order, "save")
	s.credential = v
	s.present = true
	s.saves++
	return nil
}

func (s *testStore) Delete(context.Context) error {
	s.present = false
	s.credential = userauth.Credential{}
	return nil
}

func (s *testStore) Status(context.Context) (bool, error) {
	s.order = append(s.order, "status")
	return s.present, s.statusErr
}

func (s *testStore) ConfirmUserPresence(context.Context) error {
	s.legacyConfirms++
	return s.confirmErr
}

func (s *testStore) EnsureInstallationKey(context.Context, string) (userauth.InstallationKey, error) {
	s.order = append(s.order, "installation-key")
	s.confirms++
	if s.confirmHook != nil {
		s.confirmHook()
	}
	return testInstallationKey(), s.confirmErr
}

func (s *testStore) SignInstallationProof(context.Context, userauth.InstallationKey, userauth.InstallationProof) (string, error) {
	s.order = append(s.order, "installation-proof")
	s.signs++
	return "MAYCAQECAQE", s.signErr
}

func testInstallationKey() userauth.InstallationKey {
	return userauth.InstallationKey{Algorithm: "p256-sha256-x962-der-v1", PublicKey: "BGsX0fLhLEJH-Lzm5WOkQPJ3A32BLeszoPShOUXYmMKWT-NC4v4af5uO5-tKfA-eFivOM1drMV7Oy7ZAaDe_UfU", KeyFingerprint: "698bea63dc44a344663ff1429aea10842df27b6b991ef25866b2c6c02cdcc5be"}
}

func testLoginService(t *testing.T, store *testStore, handler http.HandlerFunc) (*Service, *int) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/cli/installation-intents" {
			store.order = append(store.order, r.URL.Path)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"installation_intent": "ERERERERERERERERERERERERERERERERERERERERERE", "challenge": "IiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiI", "expires_at": time.Now().Add(10 * time.Minute).Unix(), "key_fingerprint": testInstallationKey().KeyFingerprint})
			return
		}
		handler(w, r)
	}))
	t.Cleanup(server.Close)
	opened := 0
	service, err := New(Config{APIOrigin: server.URL, Store: store, OpenBrowser: func(ctx context.Context, address string) error {
		opened++
		parsed, _ := url.Parse(address)
		q := parsed.Query()
		request, _ := http.NewRequestWithContext(ctx, http.MethodGet, q.Get("redirect_uri")+"?state="+url.QueryEscape(q.Get("state"))+"&code=fake-code", nil)
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			return err
		}
		defer response.Body.Close()
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	return service, &opened
}

func TestLoginPreflightFailureNeverOpensBrowser(t *testing.T) {
	store := &testStore{statusErr: securecredential.ErrUnavailable}
	service, opened := testLoginService(t, store, func(w http.ResponseWriter, r *http.Request) { t.Error("network on preflight failure") })
	_, err := service.Login(context.Background())
	if !errors.Is(err, securecredential.ErrUnavailable) || *opened != 0 || store.saves != 0 {
		t.Fatal("preflight was bypassed")
	}
}

func TestFailedIdentityVerificationPreservesPreviousSession(t *testing.T) {
	store := &testStore{present: true, credential: userauth.Credential{AccessToken: "old-access", RefreshToken: "old-refresh", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour)}}
	revoked := []string{}
	service, _ := testLoginService(t, store, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/token":
			fmt.Fprint(w, `{"access_token":"new-access","refresh_token":"new-refresh","token_type":"Bearer","expires_in":3600}`)
		case "/graphql":
			w.WriteHeader(http.StatusServiceUnavailable)
		case "/oauth/revoke":
			if err := r.ParseForm(); err != nil {
				t.Error(err)
			}
			revoked = append(revoked, r.Form.Get("token"))
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	result, err := service.Login(context.Background())
	if err == nil || result.Authenticated {
		t.Fatal("failed identity was authenticated")
	}
	if !store.present || store.saves != 0 || store.credential.RefreshToken != "old-refresh" {
		t.Fatal("unverified grant replaced previous session")
	}
	if len(revoked) != 1 || revoked[0] != "new-refresh" {
		t.Fatalf("wrong grant cleanup: %d calls", len(revoked))
	}
}

func TestMapStorageErrorsDoesNotReflectHelperContent(t *testing.T) {
	for _, input := range []error{securecredential.ErrInteractionRequired, securecredential.ErrCancelled} {
		if !errors.Is(mapStorageError(input), userauth.ErrStoreLocked) {
			t.Fatal("presence error not mapped")
		}
	}
	if !errors.Is(mapStorageError(errors.New("fake-secret")), securecredential.ErrUnavailable) {
		t.Fatal("unknown helper error leaked")
	}
}

func TestLoginPartialOldGrantCleanupKeepsNewVerifiedSession(t *testing.T) {
	store := &testStore{present: true, credential: userauth.Credential{AccessToken: "old-access", RefreshToken: "old-refresh", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour)}}
	revocations := 0
	service, _ := testLoginService(t, store, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/token":
			fmt.Fprint(w, `{"access_token":"new-access","refresh_token":"new-refresh","token_type":"Bearer","expires_in":3600}`)
		case "/graphql":
			if r.Header.Get("Authorization") != "Bearer new-access" {
				t.Error("identity did not use new grant")
			}
			fmt.Fprint(w, `{"data":{"getMyAccountInfo":{"id":"new-user"}}}`)
		case "/oauth/revoke":
			revocations++
			if err := r.ParseForm(); err != nil {
				t.Error(err)
			}
			if r.Form.Get("token") != "old-refresh" {
				t.Error("cleanup revoked new session")
			}
			w.WriteHeader(http.StatusServiceUnavailable)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	result, err := service.Login(context.Background())
	if err != nil || !result.Authenticated || len(result.Warnings) != 1 || store.credential.RefreshToken != "new-refresh" || revocations != 1 {
		t.Fatal("partial cleanup discarded or misreported new session")
	}
}

func TestLogoutRemoteFailureReportsLocalDeletion(t *testing.T) {
	store := &testStore{present: true, credential: userauth.Credential{AccessToken: "fake-access", RefreshToken: "fake-refresh", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour)}}
	service, _ := testLoginService(t, store, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth/revoke" {
			t.Error("unexpected logout request")
		}
		if store.present {
			t.Error("remote revocation preceded local deletion")
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	if !errors.Is(service.Logout(context.Background()), ErrLogoutRemoteIncomplete) || store.present {
		t.Fatal("logout failure lost local deletion fact")
	}
}
