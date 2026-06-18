package cursor

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestIntegration(t *testing.T) {
	if os.Getenv("AGENTBRIDGE_INTEGRATION") != "1" {
		t.Skip("AGENTBRIDGE_INTEGRATION not set")
	}
	if _, err := exec.LookPath(DefaultExecutable); err != nil {
		t.Skipf("cursor-agent not on $PATH: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	profile := integrationProfile(t, ctx)
	workdir := t.TempDir()
	res, events := runIntegrationSession(t, ctx, profile, workdir)
	if res.Status != agentbridge.ResultCompleted {
		if cursorAuthMissing(res, events) {
			t.Skip("cursor-agent authentication missing; run cursor-agent login or set CURSOR_API_KEY")
		}
		t.Fatalf("cursor integration did not complete: %+v", res)
	}
	artifact := readIntegrationArtifact(t, workdir)
	if strings.TrimSpace(artifact) != integrationArtifactBody {
		t.Fatalf("cursor artifact content = %q, want %q", artifact, integrationArtifactBody)
	}
}
