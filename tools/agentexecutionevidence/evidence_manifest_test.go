package agentexecutionevidence

import "testing"

func TestAgentExecutionEvidenceManifest(t *testing.T) {
	manifest := loadManifest(t, manifestPath)
	docText := readText(t, humanDocPath)

	assertManifestHeader(t, manifest, docText)
	seenRisks := map[string]bool{}
	for _, ev := range manifest.LocalEvidence {
		assertLocalEvidence(t, repoRoot, ev, docText)
		seenRisks[ev.Risk] = true
	}
	for _, ev := range manifest.ExternalEvidence {
		assertExternalEvidence(t, ev, docText)
		seenRisks[ev.Risk] = true
	}
	assertRequiredRisks(t, seenRisks)

	remaining := map[string]bool{}
	for _, item := range manifest.RemainingBoundaries {
		assertRemainingBoundary(t, item)
		remaining[item.ID] = true
	}
	assertRequiredRemainingBoundaries(t, remaining)
}
