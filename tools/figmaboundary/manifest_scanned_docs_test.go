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
		"docs/migration/daemon/figma-boundary-provenance.md",
	}
	scanned = appendMarkdownDocs(t, scanned, "docs/20-domain/context-map")
	scanned = appendMarkdownDocs(t, scanned, "docs/migration/daemon/figma-boundary-provenance")
	return scanned
}

func appendMarkdownDocs(t *testing.T, scanned []string, relDir string) []string {
	t.Helper()
	err := filepath.WalkDir(filepath.Join(repoRoot(t), relDir), func(path string, d fs.DirEntry, err error) error {
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
		t.Fatalf("scan %s docs: %v", relDir, err)
	}
	return scanned
}
