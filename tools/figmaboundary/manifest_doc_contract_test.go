package figmaboundary

import (
	"path/filepath"
	"testing"
)

func TestFigmaBoundaryHumanDocContainsManifestEvidence(t *testing.T) {
	manifest := loadBoundaryManifest(t)
	humanDoc := string(mustReadFile(t, filepath.Join(repoRoot(t), boundaryHumanDocRelPath)))
	for _, want := range boundaryHumanDocEvidence(manifest.SchemaVersion) {
		requireContains(t, humanDoc, want)
	}
}

func TestFigmaAIAgentDaemonBoundaryDocsStayLinked(t *testing.T) {
	root := repoRoot(t)
	for _, rel := range boundaryLinkedDocPaths() {
		body := string(mustReadFile(t, filepath.Join(root, rel)))
		requireContains(t, body, "figma-ai-agent-daemon-boundary")
	}
}

func boundaryHumanDocEvidence(schemaVersion string) []string {
	return []string{
		schemaVersion, "RIID-4843", "RIID-4847", "RIID-4851",
		"figma-metadata-page-list-underreports-pages.v1",
		"teamswyg/riido-contracts#53", "teamswyg/riido-contracts#54",
		"`stabilized_by`", "teamswyg/riido-contracts#38",
		"teamswyg/riido-contracts#52", "432:37336", "432:46849",
		"workspace-less create", "fixture", "Bottom-up",
	}
}

func boundaryLinkedDocPaths() []string {
	return []string{
		"docs/README.md",
		"docs/20-domain/context-map.md",
		"docs/20-domain/provider-runtime.md",
		"docs/30-architecture/cli-surface.md",
		"docs/migration/daemon.md",
	}
}
