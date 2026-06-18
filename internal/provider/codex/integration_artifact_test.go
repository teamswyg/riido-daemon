package codex

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func assertCodexIntegrationArtifact(t *testing.T, expected codexIntegrationExpected) {
	t.Helper()
	artifact, err := os.ReadFile(filepath.Join(expected.workdir, expected.artifactName))
	if err != nil {
		t.Fatalf(
			"codex integration completed without writing expected artifact %q in %q: %v",
			expected.artifactName,
			expected.workdir,
			err,
		)
	}
	if strings.TrimSpace(string(artifact)) != expected.artifactBody {
		t.Fatalf("codex artifact content = %q, want %q", string(artifact), expected.artifactBody)
	}
}
