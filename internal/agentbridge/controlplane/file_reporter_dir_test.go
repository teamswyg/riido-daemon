package controlplane

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileReporterCreatesDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "reports")
	if _, err := NewFileReporter(dir); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Fatalf("report path is not a dir: %s", dir)
	}
}
