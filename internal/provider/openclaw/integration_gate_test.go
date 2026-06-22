package openclaw

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func requireOpenClawIntegration(t *testing.T) (context.Context, agentbridge.DetectResult) {
	t.Helper()
	if os.Getenv("AGENTBRIDGE_INTEGRATION") != "1" {
		t.Skip("AGENTBRIDGE_INTEGRATION not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 210*time.Second)
	t.Cleanup(cancel)
	det := requireOpenClawDetectAvailable(t, ctx)
	return ctx, det
}

func requireOpenClawDetectAvailable(
	t *testing.T,
	ctx context.Context,
) agentbridge.DetectResult {
	t.Helper()
	// Two-stage gate (audit M-8): Detect must also report Available
	// before we attempt a real prompt. This handles the "binary present
	// but unusable" case (e.g. Node version too old). See
	// docs/30-architecture/integration-matrix.md §0.
	det, err := Detect(ctx, agentbridge.DetectEnv{})
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if !det.Available {
		t.Skipf("openclaw Detect reported Available=false: %s", det.Reason)
	}
	requireOpenClawConfigUsable(t, ctx, det.Executable)
	return det
}

func requireOpenClawConfigUsable(t *testing.T, ctx context.Context, exe string) {
	t.Helper()
	cmd := exec.CommandContext(ctx, exe, configProbeArgs()...)
	out, _ := cmd.CombinedOutput()
	text := strings.ToLower(string(out))
	if strings.Contains(text, "config invalid") || strings.Contains(text, "invalid config") {
		t.Skip("openclaw config invalid; run openclaw doctor --fix")
	}
	if strings.Contains(text, "connection refused by the provider endpoint") ||
		strings.Contains(text, "failovererror") {
		t.Skip("openclaw local model backend unavailable")
	}
}

func configProbeArgs() []string {
	return []string{
		"agent",
		"--local",
		"--json",
		"--session-id",
		"riido-config-probe",
		"--message",
		"Say OK only.",
		"--timeout",
		"30",
	}
}
