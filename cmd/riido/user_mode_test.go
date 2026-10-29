package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

type cliMemoryStore struct {
	mu             sync.Mutex
	credential     userauth.Credential
	present        bool
	statusCalls    int
	confirmCalls   int
	confirmErr     error
	legacyConfirms int
}

func (s *cliMemoryStore) Lock(context.Context) (func(), error) { s.mu.Lock(); return s.mu.Unlock, nil }
func (s *cliMemoryStore) Load(context.Context) (userauth.Credential, error) {
	if !s.present {
		return userauth.Credential{}, userauth.ErrNotAuthenticated
	}
	return s.credential, nil
}

func (s *cliMemoryStore) Save(_ context.Context, v userauth.Credential) error {
	s.credential = v
	s.present = true
	return nil
}

func (s *cliMemoryStore) Delete(context.Context) error {
	s.present = false
	s.credential = userauth.Credential{}
	return nil
}

func (s *cliMemoryStore) Status(context.Context) (bool, error) {
	s.statusCalls++
	return s.present, nil
}

func (s *cliMemoryStore) ConfirmUserPresence(context.Context) error {
	s.legacyConfirms++
	return s.confirmErr
}

func (s *cliMemoryStore) EnsureInstallationKey(context.Context, string) (userauth.InstallationKey, error) {
	s.confirmCalls++
	return cliInstallationKey(), s.confirmErr
}

func (*cliMemoryStore) SignInstallationProof(context.Context, userauth.InstallationKey, userauth.InstallationProof) (string, error) {
	return "MAYCAQECAQE", nil
}

func cliInstallationKey() userauth.InstallationKey {
	return userauth.InstallationKey{Algorithm: "p256-sha256-x962-der-v1", PublicKey: "BGsX0fLhLEJH-Lzm5WOkQPJ3A32BLeszoPShOUXYmMKWT-NC4v4af5uO5-tKfA-eFivOM1drMV7Oy7ZAaDe_UfU", KeyFingerprint: "698bea63dc44a344663ff1429aea10842df27b6b991ef25866b2c6c02cdcc5be"}
}

func testUserService(t *testing.T, store *cliMemoryStore) (*standalonecli.Service, *int, *int) {
	t.Helper()
	apiCalls, opened := 0, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/cli/installation-intents":
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"installation_intent": "ERERERERERERERERERERERERERERERERERERERERERE", "challenge": "IiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiI", "expires_at": time.Now().Add(10 * time.Minute).Unix(), "key_fingerprint": cliInstallationKey().KeyFingerprint})
		case "/oauth/token":
			fmt.Fprint(w, `{"access_token":"fake-access-secret","refresh_token":"fake-refresh-secret","token_type":"Bearer","expires_in":3600,"scope":"read"}`)
		case "/oauth/revoke":
			w.WriteHeader(http.StatusOK)
		case "/graphql":
			apiCalls++
			if r.Header.Get("Authorization") != "Bearer fake-access-secret" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			var request struct {
				Query     string            `json:"query"`
				Variables map[string]string `json:"variables"`
			}
			if json.NewDecoder(r.Body).Decode(&request) != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if strings.Contains(request.Query, "RiidoCLIWhoami") {
				fmt.Fprint(w, `{"data":{"getMyAccountInfo":{"id":"user-one","username":"tester"}}}`)
			} else {
				if request.Variables["componentId"] != "work-one" || request.Variables["teamId"] != "team-one" {
					t.Error("wrong user work target")
				}
				fmt.Fprint(w, `{"data":{"getComponent":{"id":"work-one","title":"Example","keyNumber":"1","document":{"id":"doc-one","HTMLContent":"<p>body</p>"}}}}`)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	service, err := standalonecli.New(standalonecli.Config{APIOrigin: server.URL, Store: store, OpenBrowser: func(ctx context.Context, raw string) error {
		opened++
		if store.statusCalls == 0 || store.confirmCalls == 0 || store.legacyConfirms != 0 {
			t.Error("browser opened before protected storage and OS confirmation")
		}
		parsed, _ := url.Parse(raw)
		q := parsed.Query()
		callback := q.Get("redirect_uri") + "?state=" + url.QueryEscape(q.Get("state")) + "&code=fixture-code"
		request, _ := http.NewRequestWithContext(ctx, http.MethodGet, callback, nil)
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
	return service, &apiCalls, &opened
}

func TestStandaloneCLILoginWhoamiWorkAndLogout(t *testing.T) {
	store := &cliMemoryStore{}
	service, calls, opened := testUserService(t, store)
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	if err := runAuth(h.io, []string{"login", "--json"}); err != nil {
		t.Fatal(err)
	}
	if !store.present || *opened != 1 || *calls != 1 || store.confirmCalls != 1 {
		t.Fatal("login did not complete actual identity read")
	}
	h.stdout.Reset()
	if err := runWhoami(h.io, []string{"--json"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(h.stdout.String(), "user-one") {
		t.Fatal("whoami omitted actual user")
	}
	h.stdout.Reset()
	if err := runWork(h.io, []string{"get", "work-one", "--team-id", "team-one", "--json"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(h.stdout.String(), "<p>body</p>") {
		t.Fatal("work API body missing")
	}
	if len(h.daemon.requests) != 0 {
		t.Fatal("standalone mode contacted daemon")
	}
	if strings.Contains(h.stdout.String()+h.stderr.String(), "fake-access-secret") || strings.Contains(h.stdout.String()+h.stderr.String(), "fake-refresh-secret") {
		t.Fatal("CLI output exposed credential")
	}
	h.stdout.Reset()
	if err := runAuth(h.io, []string{"logout", "--json"}); err != nil {
		t.Fatal(err)
	}
	if store.present {
		t.Fatal("logout retained credential")
	}
}

func TestAgentModeNeverConstructsUserSession(t *testing.T) {
	h := newContextHarness(t, agentVars())
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
		t.Fatal("agent accessed user credential factory")
		return nil, nil
	}
	for _, args := range [][]string{{"login", "--json"}, {"logout", "--json"}} {
		if got := exitCodeOf(t, runAuth(h.io, args)); got != 4 {
			t.Fatalf("agent auth exit=%d", got)
		}
	}
	h.stdout.Reset()
	if err := runAuth(h.io, []string{"status", "--json"}); err != nil {
		t.Fatal(err)
	}
}

func TestNoInputLoginNeverOpensBrowserOrStorage(t *testing.T) {
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
		t.Fatal("noninteractive login accessed protected store")
		return nil, nil
	}
	if got := exitCodeOf(t, runAuth(h.io, []string{"login", "--no-input", "--json"})); got != 3 {
		t.Fatalf("noninteractive login exit=%d", got)
	}
}

func TestUserAuthStatusDoesNotClaimAPIVerification(t *testing.T) {
	store := &cliMemoryStore{present: true, credential: userauth.Credential{AccessToken: "fake-access-secret", RefreshToken: "fake-refresh-secret", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour)}}
	service, calls, _ := testUserService(t, store)
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
	if err := runAuth(h.io, []string{"status", "--json"}); err != nil {
		t.Fatal(err)
	}
	if *calls != 0 || store.confirmCalls != 0 || !strings.Contains(h.stdout.String(), "not_checked") || strings.Contains(h.stdout.String(), "fake-access-secret") {
		t.Fatal("status confused stored credential and verified API authentication")
	}
}

func TestStandaloneAuthHelpAccuratelyDescribesSupportedMethods(t *testing.T) {
	h := newContextHarness(t, nil)
	if err := runAuth(h.io, []string{"--help"}); err != nil {
		t.Fatal(err)
	}
	output := h.stderr.String()
	if !strings.Contains(output, "login [--browser]") || !strings.Contains(output, "--desktop and --with-token are unsupported") {
		t.Fatal("auth help advertises unsupported login methods")
	}
}

func TestProtectedStorageUnavailableHasUnsupportedExit(t *testing.T) {
	if got := userAuthError(&userauth.Error{Code: "STORAGE_ERROR"}); got.Code != contextcli.CodeUnsupported {
		t.Fatal("protected storage failure must fail closed with unsupported error")
	}
}
