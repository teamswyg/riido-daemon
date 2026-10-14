package agentexecutionevidence

import "testing"

func TestAgentExecutionEvidenceManifest(t *testing.T) {
	manifest := loadManifest(t, manifestPath)
	docText := readText(t, humanDocPath)
	evidenceCorpus := readEvidenceCorpus(t, repoRoot, manifest)

	assertManifestHeader(t, manifest, docText)
	seenRisks := map[string]bool{}
	for _, ev := range collectLocalEvidence(t, manifest, manifestPath) {
		assertLocalEvidence(t, repoRoot, ev, evidenceCorpus)
		seenRisks[ev.Risk] = true
	}
	for _, ev := range collectExternalEvidence(t, manifest, manifestPath) {
		assertExternalEvidence(t, ev, evidenceCorpus)
		seenRisks[ev.Risk] = true
	}
	assertRequiredRisks(t, seenRisks)

	remaining := map[string]bool{}
	for _, item := range collectRemainingBoundaries(t, manifest, manifestPath) {
		assertRemainingBoundary(t, item)
		remaining[item.ID] = true
	}
	assertRequiredRemainingBoundaries(t, remaining)
}
