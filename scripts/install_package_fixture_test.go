package scripts_test

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// packageEntry is one archive member; a name ending in "/" is a directory.
type packageEntry struct{ name, body string }

func helperPackage(version string) []packageEntry {
	return []packageEntry{
		{"riido", "new daemon\n"},
		{"VERSION", version + "\n"},
		{"CLI_AUTH_HELPER.json", `{"schema_version":1}` + "\n"},
		{"RiidoCLIAuthHelper.app/", ""},
		{"RiidoCLIAuthHelper.app/Contents/", ""},
		{"RiidoCLIAuthHelper.app/Contents/Info.plist", "new helper\n"},
	}
}

// writePackageAssets replaces the fixture archive and SHA256SUMS.
func writePackageAssets(t *testing.T, fixture installFixture, entries []packageEntry) {
	t.Helper()
	asset := filepath.Join(fixture.assetDir, "riido-daemon_darwin_arm64.tar.gz")
	file, err := os.Create(asset)
	if err != nil {
		t.Fatalf("create archive: %v", err)
	}
	gz := gzip.NewWriter(file)
	tw := tar.NewWriter(gz)
	for _, entry := range entries {
		header := &tar.Header{Name: entry.name, Mode: 0o644, Size: int64(len(entry.body))}
		if strings.HasSuffix(entry.name, "/") {
			header.Typeflag, header.Mode, header.Size = tar.TypeDir, 0o755, 0
		} else if entry.name == "riido" {
			header.Mode = 0o755
		}
		if err := tw.WriteHeader(header); err != nil {
			t.Fatalf("write header: %v", err)
		}
		if _, err := tw.Write([]byte(entry.body)); err != nil {
			t.Fatalf("write entry: %v", err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("close gzip: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close archive: %v", err)
	}
	sums := fmt.Sprintf("%x  riido-daemon_darwin_arm64.tar.gz\n", sha256.Sum256(readFile(t, asset)))
	if err := os.WriteFile(filepath.Join(fixture.assetDir, "SHA256SUMS"), []byte(sums), 0o644); err != nil {
		t.Fatalf("write sums: %v", err)
	}
}

func runInstaller(t *testing.T, fixture installFixture, installDir string, env ...string) (string, error) {
	t.Helper()
	cmd := exec.Command("sh", "install-riido-daemon.sh")
	cmd.Dir = "."
	cmd.Env = append(os.Environ(),
		"PATH="+fixture.binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"RIIDO_DAEMON_INSTALL_DIR="+installDir,
		"INSTALL_FIXTURE_DIR="+fixture.assetDir,
		"INSTALL_MARKER="+filepath.Join(t.TempDir(), "marker"),
		"HOME="+t.TempDir(),
		"SHELL=/bin/zsh",
	)
	cmd.Env = append(cmd.Env, env...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	return string(readFile(t, path))
}
