package scripts_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallRiidoDaemonRemovesExistingBinaryBeforeInstall(t *testing.T) {
	fixture := newInstallFixture(t)
	installDir := t.TempDir()
	marker := filepath.Join(t.TempDir(), "install-marker")
	oldBinary := filepath.Join(installDir, "riido")
	if err := os.WriteFile(oldBinary, []byte("old daemon\n"), 0o755); err != nil {
		t.Fatalf("write old daemon: %v", err)
	}
	cmd := exec.Command("sh", "install-riido-daemon.sh")
	cmd.Dir = "."
	cmd.Env = append(os.Environ(),
		"PATH="+fixture.binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"RIIDO_DAEMON_VERSION=v-test",
		"RIIDO_DAEMON_INSTALL_DIR="+installDir,
		"INSTALL_FIXTURE_DIR="+fixture.assetDir,
		"INSTALL_MARKER="+marker,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("install script failed: %v\n%s", err, string(out))
	}
	markerBytes, err := os.ReadFile(marker)
	if err != nil {
		t.Fatalf("read install marker: %v", err)
	}
	if strings.TrimSpace(string(markerBytes)) != "absent" {
		t.Fatalf("old binary was still present at install time: marker=%q output=%s", string(markerBytes), out)
	}
	installed, err := os.ReadFile(oldBinary)
	if err != nil {
		t.Fatalf("read installed daemon: %v", err)
	}
	if string(installed) != "new daemon\n" {
		t.Fatalf("installed daemon = %q", string(installed))
	}
	if !strings.Contains(string(out), "riido-daemon version: v-test") {
		t.Fatalf("install output missing version: %s", string(out))
	}
}
