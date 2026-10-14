package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeLoopFixture(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	data := `{"id":"x","owner":"test","observation":{"summary":"o"},"hypothesis":{"summary":"h"},"execution":{"summary":"e"},"evaluation":{"summary":"v"},"retrospective":{"summary":"r"},"evidence":[{"kind":"command","ref":"echo ok","proves":"ok"}]}`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}
