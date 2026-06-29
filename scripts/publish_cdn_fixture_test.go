package scripts_test

import (
	"archive/tar"
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
	}
	var sums strings.Builder
	for _, asset := range assets {
		path := filepath.Join(dist, asset)
		writeCDNArchive(t, path, version)
		sum := sha256.Sum256(readFile(t, path))
		fmt.Fprintf(&sums, "%x  %s\n", sum, asset)
	}
	if err := os.WriteFile(filepath.Join(dist, "SHA256SUMS"), []byte(sums.String()), 0o644); err != nil {
		t.Fatalf("write SHA256SUMS: %v", err)
	}
	return dist
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
