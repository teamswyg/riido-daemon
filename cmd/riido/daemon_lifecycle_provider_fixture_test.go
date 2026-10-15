package main

import (
	"path/filepath"
	"testing"
)

// Lifecycle-only tests must not probe a developer's actual provider accounts.
// Missing explicit executable pins fail closed without falling back to PATH.
func isolateDaemonLifecycleProviders(t *testing.T) {
	t.Helper()
	missing := filepath.Join(t.TempDir(), "missing-provider")
	for _, key := range []string{"RIIDO_CLAUDE_PATH", "RIIDO_CODEX_PATH", "RIIDO_OPENCLAW_PATH", "RIIDO_CURSOR_PATH"} {
		t.Setenv(key, missing)
	}
}
