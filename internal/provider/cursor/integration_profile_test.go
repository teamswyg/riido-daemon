package cursor

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func integrationProfile(t *testing.T, ctx context.Context) Profile {
	t.Helper()
	det, err := Detect(ctx, agentbridge.DetectEnv{})
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if !det.Available {
		t.Skipf("cursor Detect reported Available=false: %s", det.Reason)
	}
	if ok, reason := cursorAccountAvailable(ctx); !ok {
		t.Skip(reason)
	}
	if det.Metadata != nil && det.Metadata["profile"] != "" {
		return Profile(det.Metadata["profile"])
	}
	return ProfileRootPrint
}
