package figmaboundary

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestFigmaAIAgentDaemonBoundaryRejectsStaleEvidence(t *testing.T) {
	root := repoRoot(t)
	scanned := staleEvidenceScannedDocPaths(t)
	for _, rel := range scanned {
		body := string(mustReadFile(t, filepath.Join(root, rel)))
		for _, forbidden := range []string{
			"164-50215",
			"164:50215",
			"template list",
			"starter agent templates",
			"starter agent fixtures",
			"starter fixture",
			"starter fixtures",
			"starter-fixture",
			"template descriptions/instructions",
			"dimmed template rows",
			"onboarding template catalog",
			"agent template catalog",
		} {
			if strings.Contains(body, forbidden) {
				t.Fatalf("%s contains stale Figma/template wording %q", rel, forbidden)
			}
		}
	}
}
