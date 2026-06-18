package cursor

import (
	"os"
	"path/filepath"
	"testing"
)

func readIntegrationArtifact(t *testing.T, workdir string) string {
	t.Helper()
	artifact, err := os.ReadFile(filepath.Join(workdir, integrationArtifactName))
	if err != nil {
		t.Fatalf("cursor integration completed without expected artifact in %q: %v", workdir, err)
	}
	return string(artifact)
}
