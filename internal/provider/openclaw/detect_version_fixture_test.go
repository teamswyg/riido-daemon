package openclaw

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func detectWithFixture(t *testing.T, fixture string, exitCode int) agentbridge.DetectResult {
	t.Helper()

	exe := writeShimFromFixture(t, fixture, exitCode)
	res, _ := Detect(context.Background(), agentbridge.DetectEnv{
		EnvOverride: map[string]string{EnvOverride: exe},
	})

	return res
}
