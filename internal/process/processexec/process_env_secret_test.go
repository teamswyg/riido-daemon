//go:build !windows

package processexec

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/process"
)

// TestDeviceSecretNotInAgentEnv proves provider runtime processes never
// inherit daemon-only credentials, with or without caller overrides.
func TestDeviceSecretNotInAgentEnv(t *testing.T) {
	t.Setenv("RIIDO_DEVICE_SECRET", "rdev-secret")
	t.Setenv("RIIDO_SAAS_TOKEN", "legacy-secret")
	t.Setenv("RIIDO_TOKEN", "token-secret")
	t.Setenv("RIIDO_DEVICE_ID", "device-1")
	script := `printf "%s|%s|%s|%s" "$RIIDO_DEVICE_SECRET" "$RIIDO_SAAS_TOKEN" "$RIIDO_TOKEN" "$RIIDO_DEVICE_ID"`
	cases := map[string][]string{
		"no-overrides":     nil,
		"with-overrides":   {"RIIDO_TEST_ENV=ok"},
		"explicit-attempt": {"RIIDO_DEVICE_SECRET=from-caller"},
	}
	for name, overrides := range cases {
		t.Run(name, func(t *testing.T) {
			proc, err := New().Start(context.Background(), process.Command{
				Executable: "/bin/sh",
				Args:       []string{"-c", script},
				Env:        overrides,
			})
			if err != nil {
				t.Fatalf("Start: %v", err)
			}
			stdout := string(drainAll(proc.Stdout(), 2*time.Second))
			if status := <-proc.Exited(); status.Code != 0 {
				t.Fatalf("exit: %+v", status)
			}
			if stdout != "|||device-1" {
				t.Fatalf("daemon credentials leaked to child env: %q", stdout)
			}
		})
	}
}
