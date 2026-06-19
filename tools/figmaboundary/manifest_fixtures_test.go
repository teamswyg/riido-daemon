package figmaboundary

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestFigmaAIAgentDaemonBoundaryRejectsStaleEvidence(t *testing.T) {
	root := repoRoot(t)
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
