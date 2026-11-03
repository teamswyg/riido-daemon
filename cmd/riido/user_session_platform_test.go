package main

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func TestUserSessionPlatformPaths(t *testing.T) {
	root := t.TempDir()
	exe, config, cache := filepath.Join(root, "bin", "riido"), filepath.Join(root, "config"), filepath.Join(root, "cache")
	for _, os := range []string{"darwin", "windows"} {
		helper, locks, err := userSessionPathsForOS(os, exe, config, cache)
		wantHelper, wantLocks := filepath.Join(filepath.Dir(exe), "RiidoCLIAuthHelper.app", "Contents", "MacOS", "riido-cli-auth-helper"), filepath.Join(config, "riido-cli", "auth-locks")
		if os == "windows" {
			wantHelper, wantLocks = filepath.Join(filepath.Dir(exe), "riido-cli-auth-helper.exe"), filepath.Join(cache, "riido-cli", "auth-locks")
		}
		if err != nil || helper != wantHelper || locks != wantLocks {
			t.Fatal("platform path selection changed", os, err)
		}
	}
	for _, tc := range [][4]string{{"linux", exe, config, cache}, {"windows", "relative", config, cache}, {"windows", exe, config, "relative"}, {"darwin", exe, "", cache}} {
		if _, _, err := userSessionPathsForOS(tc[0], tc[1], tc[2], tc[3]); !errors.Is(err, securecredential.ErrUnavailable) {
			t.Fatal("unsupported or relative path accepted")
		}
	}
}

func TestUserAuthorizationLauncherBoundary(t *testing.T) {
	for _, address := range []string{"file:///tmp/secret", "javascript:alert(1)", "https://user:pass@example.test/", "http://example.test/", "https://example.test/#secret", "https://example.test/\n", "https://example.test/path with spaces", "https://example.test/\"option\"", "https://example.test/?q=\"quoted\"", "https://example.test/?q=raw space", "riido-cli-auth://cli/authorize/v1", "--help"} {
		calls := 0
		err := openUserBrowserWith(t.Context(), address, func(context.Context, string) error { calls++; return nil })
		if !errors.Is(err, standalonecli.ErrBrowserUnavailable) || calls != 0 {
			t.Fatal("unsafe browser URI reached launcher")
		}
	}
	for _, address := range []string{"https://example.test/oauth/authorize?state=a%26b&code_challenge=c", "http://127.0.0.1:12345/oauth/authorize", "http://[::1]:12345/oauth/authorize"} {
		calls := 0
		err := openUserBrowserWith(t.Context(), address, func(_ context.Context, got string) error {
			calls++
			if got != address {
				t.Fatal("URI was rewritten")
			}
			return nil
		})
		if err != nil || calls != 1 {
			t.Fatal("valid authorization URI rejected", err)
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := openUserBrowserWith(ctx, "https://example.test/", func(context.Context, string) error { t.Fatal("cancelled launch"); return nil }); !errors.Is(err, context.Canceled) {
		t.Fatal("cancellation lost")
	}
	if err := openUserBrowserWith(t.Context(), "https://example.test/", func(context.Context, string) error { return errors.New("private URI or native detail") }); !errors.Is(err, standalonecli.ErrBrowserUnavailable) {
		t.Fatal("native details escaped")
	}
}

func TestUserDesktopLauncherUsesTypedKnownEnvironment(t *testing.T) {
	for _, resource := range []string{"https://monolith.riido.io", "https://monolith.api.riido.io"} {
		r := userauth.DesktopAuthorizationRequest{Resource: resource}
		want, _ := r.LaunchURL()
		calls := 0
		if err := openUserDesktopWith(t.Context(), r, func(_ context.Context, got string) error {
			calls++
			if got != want || !strings.Contains(got, "://cli/authorize/v1?request=") {
				t.Fatal("handoff changed")
			}
			return nil
		}); err != nil || calls != 1 {
			t.Fatal("typed handoff rejected", err)
		}
	}
	r := userauth.DesktopAuthorizationRequest{Resource: "https://other.example.test"}
	if err := openUserDesktopWith(t.Context(), r, func(context.Context, string) error { t.Fatal("unknown environment launched"); return nil }); !errors.Is(err, standalonecli.ErrDesktopUnavailable) {
		t.Fatal("unknown environment accepted")
	}
}
