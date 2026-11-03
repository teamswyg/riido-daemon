package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/cliconfig"
	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func TestDesktopLoginSelectsDesktopBeforeOpeningSession(t *testing.T) {
	h := newContextHarness(t, nil)
	h.io.configDir = func() (string, error) { return t.TempDir(), nil }
	opened := false
	h.io.userSession = func(args contextcli.Args) (*standalonecli.Service, error) {
		opened = true
		if args.AuthOrigin != userauth.AuthOriginDesktop || args.AuthOriginSource != "login_method" {
			t.Fatal("Desktop method did not select its own credential slot")
		}
		return nil, userauth.ErrStoreLocked
	}
	if got := exitCodeOf(t, runAuth(h.io, []string{"login", "--desktop", "--json"})); got != 3 || !opened {
		t.Fatalf("Desktop method never reached protected login: exit=%d opened=%v", got, opened)
	}
}

func TestDesktopLoginCommandCompletesAndActivatesOnlyCapturedProfile(t *testing.T) {
	for _, scenario := range []string{"success", "preference-conflict", "app-denied"} {
		t.Run(scenario, func(t *testing.T) {
			h := newContextHarness(t, nil)
			cfg := originSettings(t, h, "browser")
			store := &cliMemoryStore{}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/oauth/token":
					fmt.Fprint(w, `{"access_token":"desktop-command-access","refresh_token":"desktop-command-refresh","token_type":"Bearer","expires_in":600,"scope":"read","auth_origin":"desktop"}`)
				case "/graphql":
					fmt.Fprint(w, `{"data":{"getMyAccountInfo":{"id":"desktop-command-user"}}}`)
				default:
					t.Error("unexpected command API", r.URL.Path)
				}
			}))
			defer server.Close()
			if err := cfg.Update(t.Context(), func(c *cliconfig.Config) error {
				p := c.Profiles["production"]
				p.APIOrigin = server.URL
				c.Profiles["production"] = p
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			h.io.userSession = func(args contextcli.Args) (*standalonecli.Service, error) {
				if args.AuthOrigin != "desktop" || args.APIOrigin != server.URL {
					t.Fatal("method/resource selection changed")
				}
				return standalonecli.New(standalonecli.Config{APIOrigin: args.APIOrigin, AuthOrigin: args.AuthOrigin, Store: store, OpenDesktop: func(ctx context.Context, request userauth.DesktopAuthorizationRequest) error {
					if scenario == "preference-conflict" {
						if err := cfg.Update(ctx, func(c *cliconfig.Config) error {
							p := c.Profiles["production"]
							p.Format = "json"
							c.Profiles["production"] = p
							return nil
						}); err != nil {
							return err
						}
					}
					code := "IiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiI"
					body := map[string]any{"installation_intent": "ERERERERERERERERERERERERERERERERERERERERERE", "challenge": code, "expires_at": time.Now().Add(time.Minute).Unix(), "key_fingerprint": cliInstallationKey().KeyFingerprint, "code": code, "state": request.State, "redirect_url": request.RedirectURI + "?" + url.Values{"code": {code}, "state": {request.State}}.Encode()}
					if scenario == "app-denied" {
						body = map[string]any{"state": request.State, "error": "access_denied"}
					}
					raw, err := json.Marshal(body)
					if err != nil {
						return err
					}
					r, err := http.NewRequestWithContext(ctx, http.MethodPost, request.RedirectURI, bytes.NewReader(raw))
					if err != nil {
						return err
					}
					r.Header.Set("Content-Type", "application/json")
					response, err := http.DefaultClient.Do(r)
					if err != nil {
						return err
					}
					defer response.Body.Close()
					if response.StatusCode != http.StatusOK {
						return fmt.Errorf("callback status %d", response.StatusCode)
					}
					return nil
				}})
			}
			err := runAuth(h.io, []string{"login", "--desktop", "--json"})
			config, loadErr := cfg.Load()
			if loadErr != nil {
				t.Fatal(loadErr)
			}
			if scenario == "app-denied" {
				if err == nil || store.present || config.Profiles["production"].AuthOrigin != "browser" {
					t.Fatal("denial changed credentials or choice")
				}
				return
			}
			var reply struct{ Data standalonecli.LoginResult }
			if err != nil || json.Unmarshal(h.stdout.Bytes(), &reply) != nil || !reply.Data.Authenticated || reply.Data.AuthOrigin != "desktop" || !store.present || store.credential.Binding.AuthOrigin != "desktop" {
				t.Fatal("Desktop command did not complete", err)
			}
			if strings.Contains(h.stdout.String(), "desktop-command-access") || strings.Contains(h.stdout.String(), "desktop-command-refresh") {
				t.Fatal("token printed")
			}
			if scenario == "preference-conflict" {
				if config.Profiles["production"].AuthOrigin != "browser" || len(reply.Data.Warnings) != 1 || !strings.Contains(reply.Data.NextAction, "--auth-origin desktop") {
					t.Fatal("CAS failure lost new credential or changed choice")
				}
			} else if config.Profiles["production"].AuthOrigin != "desktop" {
				t.Fatal("success not activated")
			}
		})
	}
}

func TestDesktopLoginAgentAndNoInputDoNotOpenCredentials(t *testing.T) {
	for _, agent := range []bool{false, true} {
		vars := map[string]string(nil)
		want := 3
		if agent {
			vars = agentVars()
			want = 4
		}
		h := newContextHarness(t, vars)
		h.io.configDir = func() (string, error) { return t.TempDir(), nil }
		h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
			t.Fatal("forbidden login opened credentials")
			return nil, nil
		}
		if got := exitCodeOf(t, runAuth(h.io, []string{"login", "--desktop", "--no-input", "--json"})); got != want {
			t.Fatalf("exit %d want %d", got, want)
		}
	}
}
