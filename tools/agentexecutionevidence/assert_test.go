package agentexecutionevidence

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func assertManifestHeader(t *testing.T, manifest evidenceManifest, docText string) {
	t.Helper()
	if manifest.SchemaVersion != "riido-agent-execution-evidence.v1" {
		t.Fatalf("schema_version = %q", manifest.SchemaVersion)
	}
	if manifest.ID != "agent-execution-risk-evidence" || manifest.RiidoTask != "RIID-4964" {
		t.Fatalf("manifest identity drifted: %+v", manifest)
	}
	expectedDoc := "docs/30-architecture/agent-execution-unresolved-design/assignment-lifecycle-fsm.md"
	if manifest.HumanDoc != expectedDoc {
		t.Fatalf("human_doc = %q", manifest.HumanDoc)
	}
	if !strings.Contains(docText, "assignment-lifecycle-evidence.riido.json") {
		t.Fatal("human doc must link the executable evidence manifest")
	}
}

func assertLocalEvidence(t *testing.T, root string, ev localEvidence, docText string) {
	t.Helper()
	if ev.Risk == "" || ev.Package == "" || ev.Test == "" || ev.Proves == "" {
		t.Fatalf("local evidence must include risk/package/test/proves: %+v", ev)
	}
	if ev.Status != "verified" {
		t.Fatalf("local evidence %s status = %q", ev.Risk, ev.Status)
	}
	assertRiskKnown(t, ev.Risk)
	if strings.Contains(ev.Package, "internal/riidoaiserver") {
		t.Fatalf("daemon evidence must not reference control-plane internals: %+v", ev)
	}
	pkgDir := strings.TrimPrefix(ev.Package, "./")
	if strings.HasPrefix(pkgDir, ".") || strings.Contains(pkgDir, "..") {
		t.Fatalf("invalid local package path: %q", ev.Package)
	}
	assertTestExists(t, filepath.Join(root, filepath.FromSlash(pkgDir)), ev.Test)
	assertDocMentionsTest(t, docText, ev.Test)
}

func assertExternalEvidence(t *testing.T, ev externalEvidence, docText string) {
	t.Helper()
	if ev.Risk == "" || ev.Repo == "" || ev.Test == "" || ev.Proves == "" {
		t.Fatalf("external evidence must include risk/repo/test/proves: %+v", ev)
	}
	if ev.Status != "verified" {
		t.Fatalf("external evidence %s status = %q", ev.Risk, ev.Status)
	}
	assertRiskKnown(t, ev.Risk)
	if !slices.Contains([]string{"riido-control-plane", "riido-contracts"}, ev.Repo) {
		t.Fatalf("external evidence repo must stay at an allowed repo boundary, got %q", ev.Repo)
	}
	if strings.Contains(ev.Test, "/") || strings.Contains(ev.Test, "internal/") {
		t.Fatalf("external evidence must not reference private package paths: %+v", ev)
	}
	assertDocMentionsTest(t, docText, ev.Test)
}

func assertRemainingBoundary(t *testing.T, item remainingBoundary) {
	t.Helper()
	if item.ID == "" || item.Owner == "" || item.CurrentHandling == "" || item.RequiredNextArtifact == "" {
		t.Fatalf("remaining boundary must be actionable: %+v", item)
	}
}
