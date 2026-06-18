package supervisor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func assertCompletedArchiveManifest(t *testing.T, runWorkdir string) {
	t.Helper()

	archive, err := os.ReadFile(filepath.Join(filepath.Dir(runWorkdir), "archive.json"))
	if err != nil {
		t.Fatalf("archive manifest not written: %v", err)
	}

	for _, want := range completedArchiveManifestFields() {
		if !strings.Contains(string(archive), want) {
			t.Fatalf("archive manifest missing %q:\n%s", want, archive)
		}
	}
}

func completedArchiveManifestFields() []string {
	return []string{
		`"schema_version": "riido-workdir-archive.v1"`,
		`"retention_mode": "keep-in-place"`,
		`"result_status": "completed"`,
	}
}
