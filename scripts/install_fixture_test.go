package scripts_test

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type installFixture struct {
	assetDir string
	binDir   string
}

func newInstallFixture(t *testing.T) installFixture {
	t.Helper()
	assetDir := t.TempDir()
	binDir := t.TempDir()
	asset := filepath.Join(assetDir, "riido-daemon_darwin_arm64.tar.gz")
	writeDaemonArchive(t, asset)
	sum := sha256.Sum256(readFile(t, asset))
	sums := fmt.Sprintf("%x  riido-daemon_darwin_arm64.tar.gz\n", sum)
	if err := os.WriteFile(filepath.Join(assetDir, "SHA256SUMS"), []byte(sums), 0o644); err != nil {
		t.Fatalf("write sums: %v", err)
	}
	writeExecutable(t, filepath.Join(binDir, "curl"), fakeCurlScript())
	writeExecutable(t, filepath.Join(binDir, "install"), fakeInstallScript())
	writeExecutable(t, filepath.Join(binDir, "uname"), fakeUnameScript())
	return installFixture{assetDir: assetDir, binDir: binDir}
}

func writeDaemonArchive(t *testing.T, path string) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create archive: %v", err)
	}
	gz := gzip.NewWriter(file)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: "riido", Mode: 0o755, Size: int64(len("new daemon\n"))}); err != nil {
		t.Fatalf("write header: %v", err)
	}
	if _, err := tw.Write([]byte("new daemon\n")); err != nil {
		t.Fatalf("write daemon: %v", err)
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
}
