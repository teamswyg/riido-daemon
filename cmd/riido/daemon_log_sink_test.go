package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDaemonLogWriterDoesNotDuplicateSameFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "daemon.log")
	stderr := openAppendFile(t, path)
	defer stderr.Close()
	file := openAppendFile(t, path)
	defer file.Close()

	if _, err := fmt.Fprint(daemonLogWriter(stderr, file), "one-line\n"); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "one-line\n" {
		t.Fatalf("log = %q, want one line", got)
	}
}

func openAppendFile(t *testing.T, path string) *os.File {
	t.Helper()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	return f
}
