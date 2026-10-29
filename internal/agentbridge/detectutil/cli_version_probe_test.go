package detectutil

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestCLIVersionProbeRejectsErrorsAndSeparatesWarnings(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture")
	}
	for _, tc := range []struct{ name, body, version, reason string }{
		{"error", "echo 'Error: failed 1.2.3' >&2; exit 1", "", "version-probe-failed"},
		{"warning", "echo 'Warning: ignored' >&2; echo 'codex-cli 0.161.0'", "codex-cli 0.161.0", ""},
		{"stderr", "echo '2.1.312 (Claude Code)' >&2", "2.1.312 (Claude Code)", ""},
		{"cursor", "echo '2026.01.23-916f423'", "2026.01.23-916f423", ""},
		{"error-zero", "echo 'Error: unsupported CLI 1.2.3'", "", "version-output-invalid"},
		{"empty", "exit 0", "", "version-output-invalid"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "cli")
			if err := os.WriteFile(path, []byte("#!/bin/sh\n"+tc.body+"\n"), 0700); err != nil {
				t.Fatal(err)
			}
			version, reason := CLIVersionProbe(context.Background(), path)
			if version != tc.version || reason != tc.reason {
				t.Fatalf("version=%q reason=%q", version, reason)
			}
		})
	}
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	if version, reason := CLIVersionProbe(ctx, "/bin/sleep"); version != "" || reason != "version-probe-timeout" {
		t.Fatalf("version=%q reason=%q", version, reason)
	}
}
