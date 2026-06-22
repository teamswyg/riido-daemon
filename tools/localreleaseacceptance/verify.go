package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func verifyInstaller(ctx context.Context, root string, fixture installFixture) (scenario, string) {
	binary := filepath.Join(fixture.installDir, "riido")
	if err := os.WriteFile(binary, []byte("old daemon\n"), 0o755); err != nil {
		return failedScenario("seed old daemon: " + err.Error()), ""
	}
	if output, err := runInstaller(ctx, root, fixture); err != nil {
		return failedScenario("installer failed: " + err.Error() + " " + output), ""
	}
	if err := assertMarkerAbsent(fixture.marker); err != nil {
		return failedScenario(err.Error()), ""
	}
	version, err := installedVersion(ctx, binary)
	if err != nil {
		return failedScenario(err.Error()), version
	}
	if !strings.Contains(version, releaseVersion) {
		return failedScenario("installed version output did not include " + releaseVersion), version
	}
	return scenario{ID: "release.fresh.install", Status: statusPassed}, version
}

func runInstaller(ctx context.Context, root string, fixture installFixture) (string, error) {
	cmd := exec.CommandContext(ctx, "sh", filepath.Join(root, "scripts/install-riido-daemon.sh"))
	cmd.Env = installEnv(fixture)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	return out.String(), cmd.Run()
}

func assertMarkerAbsent(marker string) error {
	bytes, err := os.ReadFile(marker)
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(bytes)) != "absent" {
		return errOldBinaryPresent
	}
	return nil
}
