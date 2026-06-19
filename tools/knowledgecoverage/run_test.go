package main

import (
	"path/filepath"
	"testing"
)

func TestKnowledgeCoverageRepositoryManifest(t *testing.T) {
	err := run("../..", "docs/30-architecture/executable-knowledge.riido.json", "", false, true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestKnowledgeCoverageFindsUnregisteredManualDoc(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "docs/30-architecture/known.md", "# Known\n")
	writeFixture(t, root, "docs/30-architecture/unknown.md", "# Unknown\n")
	writeFixture(t, root, "docs/30-architecture/executable-knowledge.md", "")

	m := fixtureManifest()
	docs, problems := scanDocs(root, m)
	if len(docs) != 3 {
		t.Fatalf("docs len = %d", len(docs))
	}
	if len(problems) == 0 {
		t.Fatal("expected unregistered manual doc problem")
	}
}

func TestKnowledgeCoverageAllowsZeroManualGroups(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "docs/30-architecture/executable-knowledge.md", "")
	m := fixtureManifest()
	m.ManualGroups = nil
	if problems := validateManifest(root, m); len(problems) != 0 {
		t.Fatalf("zero manual groups should be valid: %v", problems)
	}
}

func fixtureManifest() manifest {
	return manifest{
		SchemaVersion:    "riido-executable-knowledge-coverage.v1",
		ID:               "fixture",
		Title:            "Fixture",
		GeneratedDoc:     "docs/30-architecture/executable-knowledge.md",
		Workflow:         "docs/30-architecture/executable-knowledge.md",
		EvidenceArtifact: "fixture",
		ScanRoots:        []string{"docs/30-architecture"},
		ManualGroups: []manualGroup{{
			ID:           "known",
			Owner:        "test",
			Reason:       "fixture",
			NextArtifact: "fixture",
			Paths:        []string{"docs/30-architecture/known.md"},
		}},
		Assertions: []string{"fixture assertion"},
	}
}

func writeFixture(t *testing.T, root, path, text string) {
	t.Helper()
	if err := writeText(filepath.Join(root, filepath.FromSlash(path)), text); err != nil {
		t.Fatal(err)
	}
}
