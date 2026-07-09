package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestVerifyInstallerReportsInstallerFailure(t *testing.T) {
	fixture, cleanup, err := newInstallFixture()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	repo := t.TempDir()
	writeInstallScript(t, repo, `#!/bin/sh
echo boom
exit 42
`)
	scenario, _ := verifyInstaller(t.Context(), repo, fixture)
	if scenario.Status != statusFailed || !strings.Contains(scenario.FailureSummary, "installer failed") {
		t.Fatalf("scenario=%+v", scenario)
	}
}

func TestVerifyInstallerReportsVersionProbeFailure(t *testing.T) {
	fixture, cleanup, err := newInstallFixture()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	repo := t.TempDir()
	writeInstallScript(t, repo, `#!/bin/sh
set -eu
cat > "$RIIDO_DAEMON_INSTALL_DIR/riido" <<'EOF'
#!/bin/sh
exit 7
EOF
chmod 0755 "$RIIDO_DAEMON_INSTALL_DIR/riido"
echo absent > "$INSTALL_MARKER"
`)
	scenario, _ := verifyInstaller(t.Context(), repo, fixture)
	if scenario.Status != statusFailed || scenario.FailureSummary == "" {
		t.Fatalf("scenario=%+v", scenario)
	}
}

func TestRunWritesFailureEvidenceForInvalidRelease(t *testing.T) {
	repo := fixtureRepo(t)
	api := releaseAPIServer(t, releaseBody(expectedReleaseAsset()))
	out := filepath.Join(t.TempDir(), "release.json")
	err := run(t.Context(), options{
		Repo:          repo,
		EvidenceOut:   out,
		ValidFor:      time.Hour,
		ReleaseAPIURL: api,
	})
	if err == nil || !strings.Contains(err.Error(), "release acceptance failed") {
		t.Fatalf("expected release acceptance failure, got %v", err)
	}
	var evidence evidenceFile
	if err := json.Unmarshal(readFile(t, out), &evidence); err != nil {
		t.Fatal(err)
	}
	if evidence.Status != statusFailed {
		t.Fatalf("expected failed evidence: %+v", evidence)
	}
}
