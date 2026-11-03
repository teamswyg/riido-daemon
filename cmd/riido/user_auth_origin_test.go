package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/clicompletion"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
	"github.com/teamswyg/riido-daemon/internal/userauth"

	"github.com/teamswyg/riido-daemon/internal/cliconfig"
	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func TestUserAuthOriginConfigSelection(t *testing.T) {
	h := newContextHarness(t, nil)
	dir := t.TempDir()
	h.io.configDir = func() (string, error) { return dir, nil }
	if err := runConfig(h.io, []string{"set", "auth_origin", "desktop", "--json"}); err != nil {
		t.Fatal(err)
	}
	h.stdout.Reset()
	if err := runConfig(h.io, []string{"get", "auth_origin", "--json"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ReplaceAll(h.stdout.String(), " ", ""), `"auth_origin":"desktop"`) {
		t.Fatal("selected origin not reported")
	}
	c, err := (cliconfig.Store{Dir: dir}).Load()
	if err != nil || !c.Valid() {
		t.Fatal(err)
	}
}

func TestUserAuthOriginFlagAndAgentBoundary(t *testing.T) {
	h := newContextHarness(t, nil)
	if _, err := h.io.parseArgs([]string{"--auth-origin", "desktop"}, contextcli.ArgSpec{}); err != nil {
		t.Fatal(err)
	}
	h = newContextHarness(t, agentVars())
	h.io.configDir = func() (string, error) { t.Fatal("agent opened config"); return "", nil }
	if _, err := h.io.parseArgs([]string{"--auth-origin", "desktop"}, contextcli.ArgSpec{}); err == nil || err.ExitCode() != 8 {
		t.Fatal(err)
	}
}

func TestUserAuthOriginPreferenceCAS(t *testing.T) {
	s := cliconfig.Store{Dir: t.TempDir()}
	c, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	r, err := cliconfig.Resolve(c, "production", true, func(string) string { return "" })
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Set(context.Background(), r, "auth_origin", "desktop"); err != nil {
		t.Fatal(err)
	}
	if err := s.Set(context.Background(), r, "auth_origin", "browser"); !errors.Is(err, cliconfig.ErrConflict) {
		t.Fatal("overwrote concurrent selection", err)
	}
}

func originSettings(t *testing.T, h *contextHarness, origin string) cliconfig.Store {
	t.Helper()
	dir := t.TempDir()
	h.io.configDir = func() (string, error) { return dir, nil }
	s := cliconfig.Store{Dir: dir}
	if err := s.Update(context.Background(), func(c *cliconfig.Config) error {
		c.Profiles["production"] = cliconfig.Profile{AuthOrigin: origin}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestUserAuthOriginSnapshotAndInvalidInput(t *testing.T) {
	h := newContextHarness(t, nil)
	s := originSettings(t, h, "desktop")
	a, err := h.io.parseArgs(nil, contextcli.ArgSpec{})
	if err != nil || a.AuthOrigin != "desktop" || a.AuthOriginSource != "profile" {
		t.Fatal(a, err)
	}
	if err := s.Update(context.Background(), func(c *cliconfig.Config) error {
		c.Profiles["production"] = cliconfig.Profile{AuthOrigin: "browser"}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	opened := false
	h.io.userSession = func(got contextcli.Args) (*standalonecli.Service, error) {
		opened = true
		if got.AuthOrigin != "desktop" {
			t.Fatal("reread preferences")
		}
		return nil, userauth.ErrNotAuthenticated
	}
	if _, err := h.io.openUserSession(a); !errors.Is(err, userauth.ErrNotAuthenticated) || !opened {
		t.Fatal(err)
	}
	a, err = h.io.parseArgs([]string{"--auth-origin", "desktop"}, contextcli.ArgSpec{})
	if err != nil || a.AuthOrigin != "desktop" || a.AuthOriginSource != "flag" {
		t.Fatal(a, err)
	}
	for _, v := range []string{"", "other", "Desktop", "desktop secret"} {
		if _, err := h.io.parseArgs([]string{"--auth-origin", v}, contextcli.ArgSpec{}); err == nil || err.ExitCode() != 2 {
			t.Fatal(v, err)
		}
	}
}

func TestUserAuthOriginLoginIndependentAndSelectionCAS(t *testing.T) {
	for _, tc := range []struct {
		name, concurrent string
		deny             bool
	}{
		{name: "activate browser"}, {name: "preference changes", concurrent: "preference"}, {name: "profile changes", concurrent: "profile"}, {name: "API server changes", concurrent: "api"}, {name: "denied", deny: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newContextHarness(t, nil)
			cfg := originSettings(t, h, "desktop")
			store := &cliMemoryStore{}
			if tc.deny {
				store.confirmErr = userauth.ErrStoreLocked
			}
			service, calls, _ := testUserService(t, store)
			h.io.userSession = func(a contextcli.Args) (*standalonecli.Service, error) {
				if a.AuthOrigin != "browser" {
					t.Fatal("browser login selected Desktop slot")
				}
				if tc.concurrent != "" {
					if err := cfg.Update(context.Background(), func(c *cliconfig.Config) error {
						if tc.concurrent == "profile" {
							c.ActiveProfile = "staging"
						} else {
							p := c.Profiles["production"]
							if tc.concurrent == "api" {
								p.APIOrigin = "https://different.example"
							} else {
								p.Format = "json"
							}
							c.Profiles["production"] = p
						}
						return nil
					}); err != nil {
						t.Fatal(err)
					}
				}
				return service, nil
			}
			err := runAuth(h.io, []string{"login", "--browser", "--json"})
			c, loadErr := cfg.Load()
			if loadErr != nil {
				t.Fatal(loadErr)
			}
			if tc.deny {
				if err == nil || store.present || *calls != 0 || c.Profiles["production"].AuthOrigin != "desktop" {
					t.Fatal("failed login changed selection", err)
				}
				return
			}
			if err != nil || !store.present || *calls != 1 {
				t.Fatal("login not authenticated", err)
			}
			var response struct{ Data standalonecli.LoginResult }
			if err := json.Unmarshal(h.stdout.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if !response.Data.Authenticated || response.Data.AuthOrigin != "browser" || response.Data.Profile != "production" {
				t.Fatal("login result lost selected identity origin")
			}
			if response.Data.APIOrigin != store.credential.Binding.Resource {
				t.Fatal("result lost login resource")
			}
			if tc.concurrent == "api" && (c.Profiles["production"].APIOrigin != "https://different.example" || response.Data.APIOrigin == c.Profiles["production"].APIOrigin) {
				t.Fatal("concurrent resource or stored grant was overwritten")
			}
			if tc.concurrent != "" {
				if c.Profiles["production"].AuthOrigin != "desktop" || !slices.Contains(response.Data.Warnings, "auth_origin_selection_not_saved") || response.Data.NextAction != "RIIDO_CLI_API_URL을 이 응답의 api_origin으로 지정한 후 riido whoami --profile production --auth-origin browser 를 실행하세요" {
					t.Fatal("concurrent selection overwritten or success hidden")
				}
			} else if c.Profiles["production"].AuthOrigin != "browser" || len(response.Data.Warnings) != 0 {
				t.Fatal("successful login did not select browser")
			}
		})
	}
}

func TestUserAuthOriginLoginRejectsReadSelector(t *testing.T) {
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
		t.Fatal("invalid login opened store")
		return nil, nil
	}
	if code := exitCodeOf(t, runAuth(h.io, []string{"login", "--auth-origin", "browser"})); code != 2 {
		t.Fatal(code)
	}
}

func TestUserAuthOriginActualHTTPReadAndLogout(t *testing.T) {
	for _, selected := range []string{"browser", "desktop"} {
		t.Run(selected, func(t *testing.T) {
			h := newContextHarness(t, nil)
			cfg := originSettings(t, h, selected)
			other := "desktop"
			if selected == other {
				other = "browser"
			}
			calls, revoke := []string{}, []string{}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/graphql":
					token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
					calls = append(calls, token)
					if token != selected {
						t.Error("HTTP used another origin")
					}
					fmt.Fprintf(w, `{"data":{"getMyAccountInfo":{"id":%q,"username":"fixture"}}}`, token)
				case "/oauth/revoke":
					if err := r.ParseForm(); err != nil {
						t.Error(err)
					}
					revoke = append(revoke, r.Form.Get("token"))
					w.WriteHeader(http.StatusOK)
				default:
					w.WriteHeader(http.StatusNotImplemented)
				}
			}))
			defer server.Close()
			stores := map[string]*cliMemoryStore{}
			for _, o := range []string{"browser", "desktop"} {
				stores[o] = &cliMemoryStore{present: true, credential: userauth.Credential{SessionID: o + "-session", AccessToken: o, RefreshToken: o + "-refresh", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour)}}
			}
			h.io.userSession = func(a contextcli.Args) (*standalonecli.Service, error) {
				// Preferences can change after parsing without changing this invocation.
				if err := cfg.Update(context.Background(), func(c *cliconfig.Config) error {
					p := c.Profiles["production"]
					p.AuthOrigin = other
					c.Profiles["production"] = p
					return nil
				}); err != nil {
					t.Fatal(err)
				}
				return standalonecli.New(standalonecli.Config{APIOrigin: server.URL, AuthOrigin: a.AuthOrigin, Store: stores[a.AuthOrigin]})
			}
			if err := runWhoami(h.io, []string{"--json"}); err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(calls, []string{selected}) || len(revoke) != 0 {
				t.Fatal("unexpected read receipt")
			}
			h.stdout.Reset()
			if err := runAuth(h.io, []string{"status", "--auth-origin", selected, "--json"}); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(strings.ReplaceAll(h.stdout.String(), " ", ""), `"auth_origin":"`+selected+`"`) || !strings.Contains(h.stdout.String(), "not_checked") || len(calls) != 1 {
				t.Fatal("status claimed verification or read another slot")
			}
			h.stdout.Reset()
			if err := runAuth(h.io, []string{"logout", "--auth-origin", selected, "--json"}); err != nil {
				t.Fatal(err)
			}
			if stores[selected].present || !stores[other].present || !slices.Equal(revoke, []string{selected + "-refresh"}) {
				t.Fatal("logout crossed selected origin")
			}
			// An absent selected slot must not reuse the still-authenticated other slot.
			if code := exitCodeOf(t, runWhoami(h.io, []string{"--auth-origin", selected, "--json"})); code != 3 || len(calls) != 1 {
				t.Fatal("credential fallback", code)
			}
		})
	}
}

func TestUserAuthOriginCompletion(t *testing.T) {
	for _, words := range [][]string{{"whoami", "--auth-origin", ""}, {"config", "set", "auth_origin", ""}} {
		if got := clicompletion.Candidates(completionCatalog(), words); !slices.Contains(got, "browser") || !slices.Contains(got, "desktop") {
			t.Fatal(got)
		}
	}
	if got := clicompletion.Candidates(completionCatalog(), []string{"auth", "login", "--auth-origin"}); len(got) != 0 {
		t.Fatal("offered read selector for login", got)
	}
}

func TestUserAuthOriginOneShotScopeIsolation(t *testing.T) {
	for _, stored := range []string{"", "browser", "desktop"} {
		t.Run(stored, func(t *testing.T) {
			selected := "desktop"
			if stored == selected {
				selected = "browser"
			}
			env := map[string]string{}
			h := newContextHarness(t, env)
			s := originSettings(t, h, stored)
			if err := s.Update(context.Background(), func(c *cliconfig.Config) error {
				p := c.Profiles["production"]
				p.SelectionOrigin = "https://monolith.riido.io"
				p.Workspace = "old-ws"
				p.Team = "old-team"
				c.Profiles["production"] = p
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			spec := contextcli.ArgSpec{UseWorkspace: true, Values: []string{"--team-id"}}
			effective := stored
			if effective == "" {
				effective = "browser"
			}
			same, sameErr := h.io.parseArgs([]string{"--auth-origin", effective}, spec)
			if sameErr != nil || same.Workspace != "old-ws" || same.Value("--team-id") != "old-team" {
				t.Fatal("same-origin defaults lost", sameErr)
			}
			a, err := h.io.parseArgs([]string{"--auth-origin", selected}, spec)
			if err != nil || a.Workspace != "" || a.Value("--team-id") != "" {
				t.Fatal("cross-origin stored scope reused", a, err)
			}
			env["RIIDO_CLI_WORKSPACE"], env["RIIDO_CLI_TEAM"] = "explicit-ws", "explicit-team"
			a, err = h.io.parseArgs([]string{"--auth-origin", selected}, spec)
			if err != nil || a.Workspace != "explicit-ws" || a.Value("--team-id") != "explicit-team" {
				t.Fatal("explicit scope lost", a, err)
			}
		})
	}
}

func TestUserAuthOriginScopePersistenceRejectsBeforeCredentials(t *testing.T) {
	h := newContextHarness(t, nil)
	originSettings(t, h, "browser")
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
		t.Fatal("mismatched scope persistence opened credentials")
		return nil, nil
	}
	for _, run := range []func() error{
		func() error {
			return runWorkspace(h.io, []string{"use", "ws-one", "--auth-origin", "desktop", "--json"})
		},
		func() error {
			return runConfig(h.io, []string{"set", "workspace", "ws-one", "--auth-origin", "desktop", "--json"})
		},
	} {
		if code := exitCodeOf(t, run()); code != int(contextcli.ExitConflict) {
			t.Fatal(code)
		}
	}
}
