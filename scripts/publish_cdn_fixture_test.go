package scripts_test

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newCDNDistFixture(t *testing.T, version string) string {
	t.Helper()
	dist := t.TempDir()
	assets := []string{
		"riido-daemon_darwin_arm64.tar.gz",
		"riido-daemon_darwin_amd64.tar.gz",
		"riido-daemon_windows_amd64.zip",
		"riido-daemon_windows_arm64.zip",
	}
	var sums strings.Builder
	for _, asset := range assets {
		path := filepath.Join(dist, asset)
		writeCDNAsset(t, path, version)
		sum := sha256.Sum256(readFile(t, path))
		fmt.Fprintf(&sums, "%x  %s\n", sum, asset)
	}
	if err := os.WriteFile(filepath.Join(dist, "SHA256SUMS"), []byte(sums.String()), 0o644); err != nil {
		t.Fatalf("write SHA256SUMS: %v", err)
	}
	return dist
}

func writeCDNAsset(t *testing.T, path, version string) {
	t.Helper()
	if strings.HasSuffix(path, ".zip") {
		writeCDNZip(t, path, version)
		return
	}
	writeCDNArchive(t, path, version)
}

func writeCDNArchive(t *testing.T, path, version string) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create CDN archive: %v", err)
	}
	gz := gzip.NewWriter(file)
	tw := tar.NewWriter(gz)
	writeTarFile(t, tw, "riido", "daemon\n", 0o755)
	writeTarFile(t, tw, "VERSION", version+"\n", 0o644)
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("close gzip: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close CDN archive: %v", err)
	}
}

func writeCDNZip(t *testing.T, path, version string) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create CDN zip: %v", err)
	}
	zw := zip.NewWriter(file)
	writeZipFile(t, zw, "riido.exe", "daemon\n")
	writeZipFile(t, zw, "VERSION", version+"\n")
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close CDN zip: %v", err)
	}
}

func writeZipFile(t *testing.T, zw *zip.Writer, name, body string) {
	t.Helper()
	file, err := zw.Create(name)
	if err != nil {
		t.Fatalf("create %s in zip: %v", name, err)
	}
	if _, err := file.Write([]byte(body)); err != nil {
		t.Fatalf("write %s in zip: %v", name, err)
	}
}

func writeTarFile(t *testing.T, tw *tar.Writer, name, body string, mode int64) {
	t.Helper()
	if err := tw.WriteHeader(&tar.Header{Name: name, Mode: mode, Size: int64(len(body))}); err != nil {
		t.Fatalf("write %s header: %v", name, err)
	}
	if _, err := tw.Write([]byte(body)); err != nil {
		t.Fatalf("write %s body: %v", name, err)
	}
}

func fakeAWSScript() string {
	return `#!/bin/sh
printf '%s\n' "$*" >> "$AWS_LOG"
`
}
