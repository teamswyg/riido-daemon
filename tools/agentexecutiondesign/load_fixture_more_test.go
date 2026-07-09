package main

import (
	"path/filepath"
	"testing"
)

func writeDesignFixtureRepo(t *testing.T, repo, manifestPath string, m model) {
	t.Helper()
	base := manifestBase(manifestPath)
	refs := fragmentRefs{
		Overview:       "overview.riido.json",
		RiskModel:      "risk.riido.json",
		ExecutionModel: "execution.riido.json",
		LifecycleModel: "lifecycle.riido.json",
		Governance:     "governance.riido.json",
	}
	m.Manifest.Fragments = refs
	m.Manifest.EvidenceManifest = "docs/evidence/manifest.riido.json"
	if err := writeJSON(repoPath(repo, manifestPath), m.Manifest); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	writeFragment(t, repo, base, refs.Overview, m.Overview)
	writeFragment(t, repo, base, refs.RiskModel, m.Risk)
	writeFragment(t, repo, base, refs.ExecutionModel, m.Execution)
	writeFragment(t, repo, base, refs.LifecycleModel, m.Lifecycle)
	writeFragment(t, repo, base, refs.Governance, m.Governance)
	writeEvidenceFixture(t, repo, m)
}

func writeFragment(t *testing.T, repo, base, rel string, value any) {
	t.Helper()
	if err := writeJSON(repoPath(repo, filepath.Join(base, rel)), value); err != nil {
		t.Fatalf("write fragment %s: %v", rel, err)
	}
}

func writeEvidenceFixture(t *testing.T, repo string, m model) {
	t.Helper()
	evidenceBase := "docs/evidence"
	ev := evidenceManifest{
		SchemaVersion: "riido-agent-execution-evidence.v1",
		ID:            "agent-execution-evidence-test",
		RiidoTask:     "RIID-test",
		HumanDoc:      "human.md",
		EvidenceFiles: evidenceFiles{
			Local:               []string{"local.json"},
			External:            []string{"external.json"},
			RemainingBoundaries: []string{"boundaries.json"},
		},
	}
	if err := writeJSON(repoPath(repo, "docs/evidence/manifest.riido.json"), ev); err != nil {
		t.Fatalf("write evidence manifest: %v", err)
	}
	items := m.Items
	if len(items) > 1 {
		items = items[:1]
	}
	if err := writeJSON(repoPath(repo, filepath.Join(evidenceBase, "local.json")), items); err != nil {
		t.Fatalf("write local evidence: %v", err)
	}
	if err := writeJSON(repoPath(repo, filepath.Join(evidenceBase, "external.json")), []evidenceItem{}); err != nil {
		t.Fatalf("write external evidence: %v", err)
	}
	if err := writeJSON(repoPath(repo, filepath.Join(evidenceBase, "boundaries.json")), m.Boundaries); err != nil {
		t.Fatalf("write boundaries: %v", err)
	}
}
