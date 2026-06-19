package main

import (
	"os"
	"testing"
)

func mustWrite(t *testing.T, path, text string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustLoad(t *testing.T, path string) manifest {
	t.Helper()
	out, err := loadManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	return out
}
