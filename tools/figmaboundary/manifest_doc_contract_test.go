package figmaboundary

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestFigmaBoundaryHumanDocContainsManifestEvidence(t *testing.T) {
	manifest := loadBoundaryManifest(t)
	humanDoc := string(mustReadFile(t, filepath.Join(repoRoot(t), boundaryHumanDocRelPath)))
	for _, want := range boundaryHumanDocEvidence(manifest) {
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

func TestFigmaBoundaryHumanDocDropsLegacyManualHardeningTasks(t *testing.T) {
	humanDoc := string(mustReadFile(t, filepath.Join(repoRoot(t), boundaryHumanDocRelPath)))
	for _, stale := range legacyBoundaryHumanDocEvidence() {
		if contains := strings.Contains(humanDoc, stale); contains {
			t.Fatalf("human doc contains stale manual hardening task %q", stale)
		}
	}
}

func boundaryHumanDocEvidence(manifest boundaryManifest) []string {
	evidence := []string{
		manifest.SchemaVersion,
		manifest.RiidoTask,
		manifest.SourceCoverageManifestProvenance.ID,
		manifest.SourceCoverageManifestProvenance.SourceFieldIntroducedBy,
		manifest.SourceCoverageManifestProvenance.MirrorsSourceField,
		"figma-metadata-page-list-underreports-pages.v1",
		"432:37336",
		"432:46849",
		"No onboarding draft execution ownership",
		"fixture",
		"Bottom-up",
	}
	evidence = append(evidence, manifest.HardeningTasks...)
	evidence = append(evidence, manifest.SourceCoverageManifestProvenance.StabilizedBy...)
	return evidence
}

func legacyBoundaryHumanDocEvidence() []string {
	return []string{
		"RIID-4847",
		"RIID-4851",
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
