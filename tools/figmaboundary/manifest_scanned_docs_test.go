package figmaboundary

import (
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

func staleEvidenceScannedDocPaths(t *testing.T) []string {
	t.Helper()
	scanned := []string{
		"docs/README.md",
		"docs/20-domain/context-map.md",
		"docs/20-domain/provider-runtime.md",
		"docs/30-architecture/cli-surface.md",
		"docs/30-architecture/figma-ai-agent-daemon-boundary.md",
		"docs/30-architecture/figma-ai-agent-daemon-boundary.riido.json",
		"docs/30-architecture/figma-ai-agent-daemon-boundary/entries.riido.json",
		"docs/migration/daemon.md",
	}
	contextMapDir := filepath.Join(repoRoot(t), "docs/20-domain/context-map")
	err := filepath.WalkDir(contextMapDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return err
		}
		rel, relErr := filepath.Rel(repoRoot(t), path)
		if relErr != nil {
			return relErr
		}
		scanned = append(scanned, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatalf("scan context-map docs: %v", err)
	}
	return scanned
}
