package detectutil

import (
	"context"
	"regexp"
	"strings"
)

var (
	cliVersionLine = regexp.MustCompile(`^(?:(?:codex-cli|claude|claude-code|cursor-agent)\s+)?v?[0-9]+\.[0-9]+\.[0-9]+(?:[-+][A-Za-z0-9.-]+)?(?:\s+\(Claude Code\))?$`)
	versionANSI    = regexp.MustCompile("\x1b\\[[0-9;]*[mK]")
)

// CLIVersionProbe accepts only a successful version line, never arbitrary CLI
// output. Diagnostics are finite codes; stderr stays local to this probe.
func CLIVersionProbe(ctx context.Context, exe string) (string, string) {
	probeCtx, cancel := context.WithTimeout(ctx, versionProbeTimeout)
	defer cancel()
	result := VersionProbeStrict(probeCtx, exe, "--version")
	if probeCtx.Err() != nil {
		return "", "version-probe-timeout"
	}
	if !result.OK || result.ExitCode != 0 {
		return "", "version-probe-failed"
	}
	for _, output := range []string{result.Stdout, result.Stderr} {
		for line := range strings.SplitSeq(output, "\n") {
			line = strings.TrimSpace(versionANSI.ReplaceAllString(line, ""))
			if len(line) <= 128 && cliVersionLine.MatchString(line) {
				return line, ""
			}
		}
	}
	return "", "version-output-invalid"
}
