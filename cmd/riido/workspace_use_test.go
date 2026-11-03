package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/cliconfig"
	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/lock"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func TestWorkspaceUseAuthenticatedAndPersistsAfterSuccess(t *testing.T) {
	for _, configAlias := range []bool{false, true} {
		t.Run(fmt.Sprint(configAlias), func(t *testing.T) {
			var calls atomic.Int32
			h := newContextHarness(t, nil)
			dir := t.TempDir()
			h.io.configDir = func() (string, error) { return dir, nil }
			store := cliconfig.Store{Dir: dir}
			service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				var body struct {
					Query     string
					Variables map[string]any
				}
				if r.Header.Get("Authorization") != "Bearer fixture-access" || json.NewDecoder(r.Body).Decode(&body) != nil || !strings.Contains(body.Query, "getCLIWorkspaces") || body.Variables["limit"] != float64(100) {
					t.Error("wrong authenticated membership read")
				}
				if _, err := os.Stat(filepath.Join(dir, "config.json")); !os.IsNotExist(err) {
					t.Error("saved before response")
				}
				fmt.Fprint(w, `{"data":{"getCLIWorkspaces":{"schemaVersion":"riido-cli-workspaces.v1","userId":"user-one","limit":100,"items":[{"id":"ws-one","name":"Workspace","isDemo":false}],"nextCursor":null,"complete":true}}}`)
			})
			h.io.userSession = func(a contextcli.Args) (*standalonecli.Service, error) {
				if a.APIOrigin != "https://monolith.api.riido.io" {
					t.Fatal(a.APIOrigin)
				}
				return service, nil
			}
			var err error
			if configAlias {
				err = runConfig(h.io, []string{"set", "workspace", "ws-one", "--profile", "staging", "--json"})
			} else {
				err = runWorkspace(h.io, []string{"use", "ws-one", "--profile", "staging", "--json"})
			}
			if err != nil || calls.Load() != 1 {
				t.Fatal(err, calls.Load())
			}
			c, err := store.Load()
			if err != nil || c.ActiveProfile != "production" || c.Profiles["staging"].Workspace != "ws-one" || c.Profiles["staging"].SelectionOrigin != "https://monolith.api.riido.io" {
				t.Fatal(c, err)
			}
			if strings.Contains(h.stdout.String(), "fixture-access") || len(h.daemon.requests) != 0 {
				t.Fatal("credential or daemon leaked")
			}
		})
	}
}

func TestConfigProfileSaveReportsNewProfile(t *testing.T) {
	h := newContextHarness(t, nil)
	dir := t.TempDir()
	h.io.configDir = func() (string, error) { return dir, nil }
	if err := runConfig(h.io, []string{"set", "profile", "staging", "--json"}); err != nil {
		t.Fatal(err)
	}
	env := decodeEnvelope(t, h.stdout.Bytes())
	raw, _ := json.Marshal(env.Data)
	var data struct{ Profile string }
	if json.Unmarshal(raw, &data) != nil || data.Profile != "staging" {
		t.Fatal("save reported previous profile", string(raw))
	}
}

func TestConfigLockTimeoutIsUnavailable(t *testing.T) {
	h := newContextHarness(t, nil)
	dir := t.TempDir()
	h.io.configDir = func() (string, error) { return dir, nil }
	guard, err := lock.AcquireFile(context.Background(), filepath.Join(dir, "config.lock"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := guard.Release(); err != nil {
			t.Error(err)
		}
	}()
	if code := exitCodeOf(t, runConfig(h.io, []string{"set", "format", "json", "--timeout", "1ms", "--json"})); code != int(contextcli.ExitUnavailable) {
		t.Fatal("timeout classified as input error", code)
	}
	if _, err := os.Stat(filepath.Join(dir, "config.json")); !os.IsNotExist(err) {
		t.Fatal("timed out command wrote config")
	}
}

func TestWorkspaceUseDoesNotSaveDeniedOrConcurrentSettings(t *testing.T) {
	for _, conflict := range []bool{false, true} {
		t.Run(fmt.Sprint(conflict), func(t *testing.T) {
			h := newContextHarness(t, nil)
			dir := t.TempDir()
			h.io.configDir = func() (string, error) { return dir, nil }
			store := cliconfig.Store{Dir: dir}
			service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
				if !conflict {
					fmt.Fprint(w, `{"errors":[{"message":"sensitive","extensions":{"code":"permission_denied"}}]}`)
					return
				}
				if err := store.Update(context.Background(), func(c *cliconfig.Config) error {
					c.Profiles["production"] = cliconfig.Profile{APIOrigin: "https://changed.example"}
					return nil
				}); err != nil {
					t.Error(err)
				}
				fmt.Fprint(w, `{"data":{"getCLIWorkspaces":{"schemaVersion":"riido-cli-workspaces.v1","userId":"user-one","limit":100,"items":[{"id":"ws-one","name":"Workspace","isDemo":false}],"nextCursor":null,"complete":true}}}`)
			})
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			want := 4
			if conflict {
				want = int(contextcli.ExitConflict)
			}
			if code := exitCodeOf(t, runWorkspace(h.io, []string{"use", "ws-one", "--json"})); code != want {
				t.Fatal(code)
			}
			c, err := store.Load()
			if err != nil || c.Profiles["production"].Workspace != "" {
				t.Fatal("saved after failure", c, err)
			}
		})
	}
}

func TestConfigCommandsAreLocalAndSanitizeRejectedValues(t *testing.T) {
	h := newContextHarness(t, nil)
	dir := t.TempDir()
	h.io.configDir = func() (string, error) { return dir, nil }
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
		t.Fatal("credentials opened for local config")
		return nil, nil
	}
	for _, raw := range [][]string{{"set", "profile", "local"}, {"set", "api_origin", "http://127.0.0.1:12345"}, {"set", "timeout", "42s"}, {"get", "timeout"}, {"list"}, {"set", "workspace", ""}} {
		if err := runConfig(h.io, append(raw, "--json")); err != nil {
			t.Fatal(raw, err)
		}
	}
	for _, raw := range [][]string{{"set", "access_token", "synthetic-secret"}, {"set", "api_origin", "https://synthetic-secret@example.com"}, {"set", "timeout", "synthetic-secret"}, {"set", "profile", "local", "synthetic-secret"}, {"set", "--token", "synthetic-secret"}} {
		if code := exitCodeOf(t, runConfig(h.io, append(raw, "--json"))); code != 2 {
			t.Fatal(raw, code)
		}
	}
	if strings.Contains(h.stdout.String(), "synthetic-secret") || strings.Contains(h.stderr.String(), "synthetic-secret") {
		t.Fatal("input secret echoed")
	}
	c, err := (cliconfig.Store{Dir: dir}).Load()
	if err != nil || c.ActiveProfile != "local" || c.Profiles["local"].APIOrigin != "http://127.0.0.1:12345" || c.Profiles["local"].Timeout != "42s" {
		t.Fatal(c, err)
	}
}

func TestConfigAndWorkspaceUseRejectAgentBeforeSettings(t *testing.T) {
	h := newContextHarness(t, map[string]string{"RIIDO_RUN_TOKEN": "fixture"})
	h.io.configDir = func() (string, error) { t.Fatal("agent opened config"); return "", nil }
	if code := exitCodeOf(t, runConfig(h.io, []string{"list", "--json"})); code != 8 {
		t.Fatal(code)
	}
	if code := exitCodeOf(t, runWorkspace(h.io, []string{"use", "ws-one", "--json"})); code != 8 {
		t.Fatal(code)
	}
	if len(h.daemon.requests) != 0 {
		t.Fatal("sent daemon request")
	}
}
