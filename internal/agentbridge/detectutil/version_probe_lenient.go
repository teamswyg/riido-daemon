package detectutil

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strings"
)

// VersionProbe runs `<exe> <args...>` with a short timeout and returns the
// trimmed combined stdout+stderr.
func VersionProbe(ctx context.Context, exe string, args ...string) (string, bool) {
	if exe == "" {
		return "", false
	}
	probeCtx, cancel := context.WithTimeout(ctx, versionProbeTimeout)
	defer cancel()
	cmd := exec.CommandContext(probeCtx, exe, args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		if errors.Is(probeCtx.Err(), context.DeadlineExceeded) || buf.Len() == 0 {
			return "", false
		}
	}
	return strings.TrimSpace(buf.String()), true
}
