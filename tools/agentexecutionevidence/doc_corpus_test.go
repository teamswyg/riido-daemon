package agentexecutionevidence

import (
	"path/filepath"
	"strings"
	"testing"
)

func readEvidenceCorpus(t *testing.T, root string, manifest evidenceManifest) string {
	t.Helper()
	seen := map[string]bool{}
	paths := append([]string{manifest.HumanDoc}, manifest.SourceDocuments...)
	parts := make([]string, 0, len(paths))
	for _, path := range paths {
		if seen[path] {
			continue
		}
		seen[path] = true
		parts = append(parts, readText(t, filepath.Join(root, filepath.FromSlash(path))))
	}
	return strings.Join(parts, "\n")
}
