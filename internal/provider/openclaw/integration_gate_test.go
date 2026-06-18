package openclaw

import (
	"context"
	"os"
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
	return det
}
