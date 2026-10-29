package codex

import (
	"context"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
)

func detectedVersion(ctx context.Context, executable string) (string, bool) {
	probe := detectutil.VersionProbeStrict(ctx, executable, "--version")
	version := strings.TrimSpace(probe.Stdout)
	return version, probe.OK && probe.ExitCode == 0 && version != ""
}
