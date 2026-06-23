package codex

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func requireCodexIntegration(t *testing.T) context.Context {
	t.Helper()
	if os.Getenv("AGENTBRIDGE_INTEGRATION") != "1" {
		t.Skip("AGENTBRIDGE_INTEGRATION not set")
	}
	if _, err := exec.LookPath(DefaultExecutable); err != nil {
		t.Skipf("codex not on $PATH: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), codexIntegrationContextTimeout)
	t.Cleanup(cancel)
	requireCodexDetectAvailable(t, ctx)
	return ctx
}

func requireCodexDetectAvailable(t *testing.T, ctx context.Context) {
	t.Helper()
	// Two-stage gate (audit M-8 / integration-matrix.md §0).
	det, err := Detect(ctx, agentbridge.DetectEnv{})
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if !det.Available {
		t.Skipf("codex Detect reported Available=false: %s", det.Reason)
	}
}
