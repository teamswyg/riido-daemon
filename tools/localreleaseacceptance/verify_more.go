package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
)

var errOldBinaryPresent = errors.New("old daemon was present during install copy")

func installEnv(fixture installFixture) []string {
	path := fixture.binDir + string(os.PathListSeparator) + os.Getenv("PATH")
	return append(os.Environ(),
		"PATH="+path,
		"RIIDO_DAEMON_VERSION="+releaseVersion,
		"RIIDO_DAEMON_INSTALL_DIR="+fixture.installDir,
		"INSTALL_FIXTURE_DIR="+fixture.assetDir,
		"INSTALL_MARKER="+fixture.marker,
	)
}

func installedVersion(ctx context.Context, binary string) (string, error) {
	cmd := exec.CommandContext(ctx, binary, "version")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return out.String(), err
	}
	return strings.TrimSpace(out.String()), nil
}

func failedScenario(id, summary string) scenario {
	return scenario{
		ID:             id,
		Status:         statusFailed,
		FailureSummary: summary,
	}
}
