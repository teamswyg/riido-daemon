package agentexecutionevidence

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAgentExecutionEvidenceManifest(t *testing.T) {
	root := filepath.Join("..", "..")
	manifestPath := filepath.Join(root, "docs", "30-architecture", "agent-execution-unresolved-design", "assignment-lifecycle-evidence.riido.json")
	docPath := filepath.Join(root, "docs", "30-architecture", "agent-execution-unresolved-design", "assignment-lifecycle-fsm.md")

	manifest := loadManifest(t, manifestPath)
	docText := readText(t, docPath)

	if manifest.SchemaVersion != "riido-agent-execution-evidence.v1" {
		t.Fatalf("schema_version = %q", manifest.SchemaVersion)
	}
	if manifest.ID != "agent-execution-risk-evidence" || manifest.RiidoTask != "RIID-4964" {
		t.Fatalf("manifest identity drifted: %+v", manifest)
	}
	if manifest.HumanDoc != "docs/30-architecture/agent-execution-unresolved-design/assignment-lifecycle-fsm.md" {
		t.Fatalf("human_doc = %q", manifest.HumanDoc)
	}
	if !strings.Contains(docText, "assignment-lifecycle-evidence.riido.json") {
		t.Fatal("human doc must link the executable evidence manifest")
	}

	seenRisks := map[string]bool{}
	for _, ev := range manifest.LocalEvidence {
		assertLocalEvidence(t, root, ev)
		seenRisks[ev.Risk] = true
		if !strings.Contains(docText, ev.Test) {
			t.Fatalf("human doc must mention local evidence test %q", ev.Test)
		}
	}
	for _, ev := range manifest.ExternalEvidence {
		assertExternalEvidence(t, ev)
		seenRisks[ev.Risk] = true
		if !strings.Contains(docText, ev.Test) {
			t.Fatalf("human doc must mention external evidence test %q", ev.Test)
		}
	}
	for _, risk := range []string{
		"same-task-multiple-assignments",
		"public-repo-worktree-materialization",
		"private-repo-fail-closed",
		"restart-recovery-refuses-fresh-start",
		"restart-recovery-provider-session-resume",
		"active-stream-handoff",
		"terminal-late-progress-fence",
	} {
		if !seenRisks[risk] {
			t.Fatalf("manifest missing required risk evidence %q", risk)
		}
	}

	remaining := map[string]bool{}
	for _, item := range manifest.RemainingBoundaries {
		if item.ID == "" || item.Owner == "" || item.CurrentHandling == "" || item.RequiredNextArtifact == "" {
			t.Fatalf("remaining boundary must be actionable: %+v", item)
		}
		remaining[item.ID] = true
	}
	for _, id := range []string{
		"private-repo-auth",
		"web-approval-round-trip",
		"client-desktop-consumption",
		"generated-fsm-conformance",
	} {
		if !remaining[id] {
			t.Fatalf("remaining boundary %q must stay explicit", id)
		}
	}
}

func assertLocalEvidence(t *testing.T, root string, ev localEvidence) {
	t.Helper()
	if ev.Risk == "" || ev.Package == "" || ev.Test == "" || ev.Proves == "" {
		t.Fatalf("local evidence must include risk/package/test/proves: %+v", ev)
	}
	if ev.Status != "verified" {
		t.Fatalf("local evidence %s status = %q", ev.Risk, ev.Status)
	}
	if strings.Contains(ev.Package, "internal/riidoaiserver") {
		t.Fatalf("daemon evidence must not reference control-plane internals: %+v", ev)
	}
	pkgDir := strings.TrimPrefix(ev.Package, "./")
	if strings.HasPrefix(pkgDir, ".") || strings.Contains(pkgDir, "..") {
		t.Fatalf("invalid local package path: %q", ev.Package)
	}
	assertTestExists(t, filepath.Join(root, filepath.FromSlash(pkgDir)), ev.Test)
}

func assertExternalEvidence(t *testing.T, ev externalEvidence) {
	t.Helper()
	if ev.Risk == "" || ev.Repo == "" || ev.Test == "" || ev.Proves == "" {
		t.Fatalf("external evidence must include risk/repo/test/proves: %+v", ev)
	}
	if ev.Status != "verified" {
		t.Fatalf("external evidence %s status = %q", ev.Risk, ev.Status)
	}
	if ev.Repo != "riido-control-plane" {
		t.Fatalf("external evidence repo must be named at repo boundary, got %q", ev.Repo)
	}
	if strings.Contains(ev.Test, "/") || strings.Contains(ev.Test, "internal/") {
		t.Fatalf("external evidence must not reference private package paths: %+v", ev)
	}
}

func assertTestExists(t *testing.T, dir, testName string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read package dir %s: %v", dir, err)
	}
	needle := "func " + testName + "("
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		if strings.Contains(readText(t, filepath.Join(dir, entry.Name())), needle) {
			return
		}
	}
	t.Fatalf("test %s not found under %s", testName, dir)
}

func loadManifest(t *testing.T, path string) evidenceManifest {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader([]byte(readText(t, path))))
	dec.DisallowUnknownFields()
	var manifest evidenceManifest
	if err := dec.Decode(&manifest); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	return manifest
}

func readText(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

type evidenceManifest struct {
	SchemaVersion       string              `json:"schema_version"`
	ID                  string              `json:"id"`
	RiidoTask           string              `json:"riido_task"`
	HumanDoc            string              `json:"human_doc"`
	SourceDocuments     []string            `json:"source_documents"`
	LocalEvidence       []localEvidence     `json:"local_evidence"`
	ExternalEvidence    []externalEvidence  `json:"external_evidence"`
	RemainingBoundaries []remainingBoundary `json:"remaining_boundaries"`
}

type localEvidence struct {
	Risk    string `json:"risk"`
	Status  string `json:"status"`
	Package string `json:"package"`
	Test    string `json:"test"`
	Proves  string `json:"proves"`
}

type externalEvidence struct {
	Risk   string `json:"risk"`
	Status string `json:"status"`
	Repo   string `json:"repo"`
	Test   string `json:"test"`
	Proves string `json:"proves"`
}

type remainingBoundary struct {
	ID                   string `json:"id"`
	Owner                string `json:"owner"`
	CurrentHandling      string `json:"current_handling"`
	RequiredNextArtifact string `json:"required_next_artifact"`
}
