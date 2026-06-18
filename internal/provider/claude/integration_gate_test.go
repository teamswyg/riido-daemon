package claude

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func claudeIntegrationContext(t *testing.T) context.Context {
	t.Helper()

	if os.Getenv("AGENTBRIDGE_INTEGRATION") != "1" {
		t.Skip("AGENTBRIDGE_INTEGRATION not set")
	}
	if _, err := exec.LookPath(DefaultExecutable); err != nil {
		t.Skipf("claude not on $PATH: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	t.Cleanup(cancel)
	requireClaudeIntegrationDetected(t, ctx)

	return ctx
}

func requireClaudeIntegrationDetected(t *testing.T, ctx context.Context) {
	t.Helper()

	det, err := Detect(ctx, agentbridge.DetectEnv{})
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if !det.Available {
		t.Skipf("claude Detect reported Available=false: %s", det.Reason)
	}
}
