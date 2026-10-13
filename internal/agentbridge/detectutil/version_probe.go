package detectutil

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/process/execwindow"
)

// VersionProbeStrict runs the same probe as VersionProbe but exposes
// the exit code so callers can fail closed when the CLI reports an
// error through a non-zero exit.
func VersionProbeStrict(ctx context.Context, exe string, args ...string) ProbeResult {
	if exe == "" {
		return ProbeResult{OK: false}
	}
	probeCtx, cancel := context.WithTimeout(ctx, versionProbeTimeout)
	defer cancel()
	cmd := exec.CommandContext(probeCtx, exe, args...)
	execwindow.Hide(cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if errors.Is(probeCtx.Err(), context.DeadlineExceeded) {
		return ProbeResult{OK: false}
	}
	exitCode := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			// Failed to start, or killed by a signal we can't classify.
			return ProbeResult{OK: false}
		}
	}
	return ProbeResult{
		Output:   strings.TrimSpace(stdout.String() + stderr.String()),
		Stdout:   strings.TrimSpace(stdout.String()),
		Stderr:   strings.TrimSpace(stderr.String()),
		ExitCode: exitCode,
		OK:       true,
	}
}
