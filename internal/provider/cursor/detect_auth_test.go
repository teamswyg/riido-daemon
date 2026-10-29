package cursor

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/teamswyg/riido-contracts/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestDetectUnauthenticatedSuccessfulStatusIsUnavailable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix provider fixture")
	}
	executable := filepath.Join(t.TempDir(), "cursor-agent")
	const script = "#!/bin/sh\ncase \"$1\" in\nstatus) printf 'Not logged in\\n';;\n--version) printf '2026.01.23-fixture\\n';;\n--help) printf 'Usage: agent [options]\\n';;\n*) exit 2;;\nesac\n"
	if err := os.WriteFile(executable, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	result, err := Detect(context.Background(), agentbridge.DetectEnv{
		EnvOverride: map[string]string{EnvOverride: executable},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Available || result.HealthStatus != hostintegration.ProviderHealthUnavailable ||
		result.DiagnosticCode != hostintegration.ProviderDiagnosticLoginRequired {
		t.Fatalf("unauthenticated provider classified as available=%t health=%s diagnostic=%s",
			result.Available, result.HealthStatus, result.DiagnosticCode)
	}
	if result.Executable != executable || result.Version != "2026.01.23-fixture" {
		t.Fatal("authentication state discarded executable identity")
	}
}
