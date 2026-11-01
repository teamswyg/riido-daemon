package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/cliconfig"
	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func TestUserSettingsPrecedenceAndCapturedOrigin(t *testing.T) {
	env := map[string]string{"RIIDO_CLI_PROFILE": "staging", "RIIDO_CLI_WORKSPACE": "env-ws", "RIIDO_CLI_TEAM": "env-team", "RIIDO_CLI_FORMAT": "json", "RIIDO_CLI_TIMEOUT": "45s"}
	h := newContextHarness(t, env)
	dir := t.TempDir()
	h.io.configDir = func() (string, error) { return dir, nil }
	s := cliconfig.Store{Dir: dir}
	if err := s.Update(context.Background(), func(c *cliconfig.Config) error {
		c.Profiles["staging"] = cliconfig.Profile{SelectionOrigin: "https://monolith.api.riido.io", Workspace: "saved-ws", Team: "saved-team", Format: "text", Timeout: "40s"}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	spec := contextcli.ArgSpec{UseWorkspace: true, Values: []string{"--team-id"}}
	a, e := h.io.parseArgs(nil, spec)
	if e != nil || a.Profile != "staging" || a.APIOrigin != "https://monolith.api.riido.io" || a.Workspace != "env-ws" || a.Value("--team-id") != "env-team" || !a.JSON || a.Timeout != 45*time.Second {
		t.Fatal(a, e)
	}
	env["RIIDO_CLI_API_URL"] = "https://different.example"
	if origin, _, err := h.io.userAPIOrigin(a); err != nil || origin != "https://monolith.api.riido.io" {
		t.Fatal("origin reread", origin, err)
	}
	env["RIIDO_CLI_FORMAT"], env["RIIDO_CLI_TIMEOUT"], env["RIIDO_CLI_TEAM"], env["RIIDO_CLI_WORKSPACE"] = "invalid", "invalid", "invalid space", "invalid space"
	a, e = h.io.parseArgs([]string{"--profile", "production", "--workspace", "flag-ws", "--team", "flag-team", "--format", "text", "--timeout", "12s"}, spec)
	if e != nil || a.Profile != "production" || a.Workspace != "flag-ws" || a.Value("--team-id") != "flag-team" || a.JSON || a.Timeout != 12*time.Second {
		t.Fatal(a, e)
	}
}

func TestUserSettingsTeamAliasAndScope(t *testing.T) {
	h := newContextHarness(t, nil)
	spec := contextcli.ArgSpec{Values: []string{"--team-id"}}
	for _, raw := range [][]string{{"--team", "one", "--team-id", "two"}, {"--team", ""}} {
		if _, err := h.io.parseArgs(raw, spec); err == nil || err.ExitCode() != 2 {
			t.Fatal(raw, err)
		}
	}
	a, e := h.io.parseArgs([]string{"--team", "one", "--team-id", "one"}, spec)
	if e != nil || a.Value("--team-id") != "one" {
		t.Fatal(a, e)
	}
	if _, err := h.io.parseArgs([]string{"--team", "one"}, contextcli.ArgSpec{}); err == nil || err.ExitCode() != 8 {
		t.Fatal(err)
	}
}

func TestSettingsHelpAgentAndCorruptIsolation(t *testing.T) {
	for _, agent := range []bool{false, true} {
		env := map[string]string{}
		if agent {
			env["RIIDO_RUN_TOKEN"] = "fixture"
		}
		h := newContextHarness(t, env)
		h.io.configDir = func() (string, error) { t.Fatal("read config for help/agent"); return "", nil }
		if _, err := h.io.parseArgs([]string{"--help"}, contextcli.ArgSpec{}); err != nil {
			t.Fatal(err)
		}
		if agent {
			if _, err := h.io.parseArgs(nil, contextcli.ArgSpec{}); err != nil {
				t.Fatal(err)
			}
			if _, err := h.io.parseArgs([]string{"--profile", "staging"}, contextcli.ArgSpec{}); err == nil || err.ExitCode() != 8 {
				t.Fatal(err)
			}
		}
	}
	h := newContextHarness(t, nil)
	dir := t.TempDir()
	h.io.configDir = func() (string, error) { return dir, nil }
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte("malformed"), 0o600); err != nil {
		t.Fatal(err)
	}
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
		t.Fatal("credentials opened after corrupt config")
		return nil, nil
	}
	if code := exitCodeOf(t, runWorkspace(h.io, []string{"list", "--json"})); code != 2 {
		t.Fatal(code)
	}
}

func TestUserSettingsWorkspaceOnlyWhereSupported(t *testing.T) {
	h := newContextHarness(t, map[string]string{"RIIDO_CLI_WORKSPACE": "default-ws", "RIIDO_CLI_TEAM": "default-team"})
	a, e := h.io.parseArgs(nil, contextcli.ArgSpec{})
	if e != nil || a.Workspace != "" || a.HasValue("--team-id") {
		t.Fatal("defaults leaked to unsupported command", a, e)
	}
	a, e = h.io.parseArgs(nil, contextcli.ArgSpec{UseWorkspace: true})
	if e != nil || a.Workspace != "default-ws" || !a.HasValue("--workspace") {
		t.Fatal(a, e)
	}
	a, e = h.io.parseArgs([]string{"--workspace", ""}, contextcli.ArgSpec{UseWorkspace: true})
	if e != nil || a.Workspace != "" {
		t.Fatal("empty flag lost", a, e)
	}
}

func TestLocalUnconfiguredNeverOpensCredentials(t *testing.T) {
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
		t.Fatal("fell back to production")
		return nil, nil
	}
	if code := exitCodeOf(t, runWorkspace(h.io, []string{"list", "--profile", "local", "--json"})); code != 2 {
		t.Fatal(code)
	}
}

func TestAuthTimeoutKeepsEnvironmentAndProfile(t *testing.T) {
	for _, tc := range []struct {
		env  map[string]string
		raw  []string
		want time.Duration
	}{
		{nil, nil, 5 * time.Minute},
		{map[string]string{"RIIDO_CLI_TIMEOUT": "42s"}, nil, 42 * time.Second},
		{map[string]string{"RIIDO_CLI_TIMEOUT": "42s"}, []string{"--timeout", "17s"}, 17 * time.Second},
	} {
		t.Run(tc.want.String(), func(t *testing.T) {
			h := newContextHarness(t, tc.env)
			opened := false
			h.io.userSession = func(a contextcli.Args) (*standalonecli.Service, error) {
				opened = true
				if a.Timeout != tc.want {
					t.Fatal(a.Timeout)
				}
				return nil, userapi.ErrAuthentication
			}
			if code := exitCodeOf(t, h.io.runAuthLogin(tc.raw)); code != 3 || !opened {
				t.Fatal(code, opened)
			}
		})
	}
	h := newContextHarness(t, nil)
	dir := t.TempDir()
	h.io.configDir = func() (string, error) { return dir, nil }
	s := cliconfig.Store{Dir: dir}
	if err := s.Update(context.Background(), func(c *cliconfig.Config) error {
		c.Profiles["production"] = cliconfig.Profile{Timeout: "53s"}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	h.io.userSession = func(a contextcli.Args) (*standalonecli.Service, error) {
		if a.Timeout != 53*time.Second {
			t.Fatal(a.Timeout)
		}
		return nil, userapi.ErrAuthentication
	}
	if code := exitCodeOf(t, h.io.runAuthLogin(nil)); code != 3 {
		t.Fatal(code)
	}
}
