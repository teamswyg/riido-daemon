package main

import (
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func TestDesktopDaemonSocketFollowsBuildProfile(t *testing.T) {
	home := "/Users/someone"
	base := filepath.Join(home, "Library", "Application Support", "riido-desktop")
	always := func(string) bool { return true }
	for _, tc := range []struct {
		goos, profile, want string
	}{
		{"darwin", "production", filepath.Join(base, "ai-agent-daemon", "agentd.sock")},
		{"darwin", "staging", filepath.Join(home, "Library", "Application Support", "Riido Staging", "ai-agent-daemon-staging", "agentd.sock")},
		{"darwin", "local", ""},
		{"linux", "staging", ""},
		{"windows", "production", ""},
	} {
		got, ok := desktopDaemonSocketFor(tc.goos, home, tc.profile, always)
		if tc.want == "" && ok || tc.want != "" && (!ok || got != tc.want) {
			t.Fatalf("%s/%s = %q %v, want %q", tc.goos, tc.profile, got, ok, tc.want)
		}
	}
	if _, ok := desktopDaemonSocketFor("darwin", home, "staging", func(string) bool { return false }); ok {
		t.Fatal("missing desktop socket was selected")
	}
}

func TestDaemonSocketPrefersExplicitThenDesktopInUserMode(t *testing.T) {
	io := contextIO{
		defaultSocket: func() (string, error) { return "/default.sock", nil },
		desktopSocket: func() (string, bool) { return "/desktop.sock", true },
	}
	if got, _ := io.daemonSocket(contextcli.Environment{Mode: contextcli.ModeUser, DaemonSocket: "/explicit.sock"}); got != "/explicit.sock" {
		t.Fatalf("explicit socket = %q", got)
	}
	if got, _ := io.daemonSocket(contextcli.Environment{Mode: contextcli.ModeUser}); got != "/desktop.sock" {
		t.Fatalf("user socket = %q", got)
	}
	if got, _ := io.daemonSocket(contextcli.Environment{Mode: contextcli.ModeAgent}); got != "/default.sock" {
		t.Fatalf("agent socket = %q", got)
	}
	io.desktopSocket = func() (string, bool) { return "", false }
	if got, _ := io.daemonSocket(contextcli.Environment{Mode: contextcli.ModeUser}); got != "/default.sock" {
		t.Fatalf("fallback socket = %q", got)
	}
}

func TestStagingDesktopSocketPrefersIsolatedThenLegacy(t *testing.T) {
	home := t.TempDir()
	base := filepath.Join(home, "Library", "Application Support")
	isolated := filepath.Join(base, "Riido Staging", "ai-agent-daemon-staging", "agentd.sock")
	legacy := filepath.Join(base, "riido-desktop", "ai-agent-daemon-staging", "agentd.sock")
	production := filepath.Join(base, "riido-desktop", "ai-agent-daemon", "agentd.sock")
	for _, tc := range []struct {
		name  string
		files map[string]bool
		want  string
	}{
		{"isolated", map[string]bool{isolated: true}, isolated},
		{"legacy", map[string]bool{legacy: true}, legacy},
		{"both prefer isolated", map[string]bool{isolated: true, legacy: true}, isolated},
		{"production is not staging", map[string]bool{production: true}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := desktopDaemonSocketFor("darwin", home, "staging", func(p string) bool { return tc.files[p] })
			if ok != (tc.want != "") || ok && got != tc.want {
				t.Fatalf("got %q %v, want %q", got, ok, tc.want)
			}
		})
	}
}
