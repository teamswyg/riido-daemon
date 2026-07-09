package main

import (
	"strings"
	"testing"
)

func TestRenderedDocsCoverDesignSurfaces(t *testing.T) {
	m := testModel()
	docs := renderedDocs(m)
	if len(docs) != 14 {
		t.Fatalf("docs = %d, want 14", len(docs))
	}
	assertDocContains(t, docs[m.Manifest.GeneratedDoc], "Executable SSOT", "Overview")
	assertDocContains(t, docs[baseDir+"overview.md"], "Riido task: RIID-test", "observe")
	assertDocContains(t, docs[baseDir+"problem-map.md"], "risk-one", "direction")
	assertDocContains(t, docs[baseDir+"execution-identity.md"], "assignment_id", "assignment id is stable")
	assertDocContains(t, docs[baseDir+"retry-recovery-policy.md"], "fail closed", "auth")
	assertDocContains(t, docs[baseDir+"verification-evidence.md"], "risk-one", "riido-daemon:TestOne")
	assertDocContains(t, docs[baseDir+"current-daemon-slice-status.md"], "Remaining boundaries", "verifier")
	assertDocContains(t, docs[baseDir+"rag-guardrails.md"], "public docs", "secrets")
}

func TestBuildEvidenceSortsGeneratedDocs(t *testing.T) {
	m := testModel()
	docs := map[string]string{"z.md": "z", "a.md": "a"}
	result := buildEvidence(m, docs, []string{"stale"})
	if result.Status != "failed" || result.EvidenceItems != 1 || result.Remaining != 1 {
		t.Fatalf("result = %+v", result)
	}
	if len(result.GeneratedDocs) != 2 ||
		result.GeneratedDocs[0] != "a.md" ||
		result.GeneratedDocs[1] != "z.md" {
		t.Fatalf("generated docs not sorted: %+v", result.GeneratedDocs)
	}
}

func assertDocContains(t *testing.T, body string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(body, want) {
			t.Fatalf("doc missing %q:\n%s", want, body)
		}
	}
}
