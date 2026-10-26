package codex

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/teamswyg/riido-contracts/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestDetectIncompleteProbesDenyNewWork(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix provider fixture")
	}
	for _, tt := range []struct {
		name, auth, version string
		want                hostintegration.ProviderDiagnosticCode
	}{
		{"auth unknown", "exit 7", "printf 'codex-cli 1.0.0'", hostintegration.ProviderDiagnosticAuthProbeFailed},
		{"version failure", "exit 0", "printf 'temporary failure' >&2; exit 7", hostintegration.ProviderDiagnosticVersionProbeFailed},
		{"version empty", "exit 0", "exit 0", hostintegration.ProviderDiagnosticVersionProbeFailed},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := detectWithProbeScript(t, tt.auth, tt.version)
			if result.Available || result.HealthStatus != hostintegration.ProviderHealthUnknown || result.DiagnosticCode != tt.want {
				t.Fatalf("incomplete probe admitted work: available=%t health=%s diagnostic=%s", result.Available, result.HealthStatus, result.DiagnosticCode)
			}
		})
	}
}

func TestDetectVersionIgnoresStderrWarnings(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix provider fixture")
	}
	result := detectWithProbeScript(t, "exit 0", "printf 'codex-cli 1.0.0'; printf 'startup warning' >&2")
	if !result.Available || result.Version != "codex-cli 1.0.0" {
		t.Fatalf("warning changed provider version: available=%t version=%q", result.Available, result.Version)
	}
}

func detectWithProbeScript(t *testing.T, auth, version string) agentbridge.DetectResult {
	t.Helper()
	exe := filepath.Join(t.TempDir(), "codex")
	script := "#!/bin/sh\ncase \"$1\" in\nlogin) " + auth + ";;\n--version) " + version + ";;\n*) exit 2;;\nesac\n"
	if err := os.WriteFile(exe, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	result, err := Detect(context.Background(), agentbridge.DetectEnv{EnvOverride: map[string]string{EnvOverride: exe}})
	if err != nil {
		t.Fatal(err)
	}
	return result
}
