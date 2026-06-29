package main

import (
	"path/filepath"
	"testing"
)

func fixtureManifest() manifest {
	return manifest{
		SchemaVersion:    "riido-executable-knowledge-coverage.v1",
		ID:               "fixture",
		Title:            "Fixture",
		GeneratedDoc:     "docs/30-architecture/executable-knowledge.md",
		Workflow:         "docs/30-architecture/executable-knowledge.md",
		EvidenceArtifact: "fixture",
		ScanRoots:        []string{"docs/30-architecture"},
		LoopRegistry: []loopRegistryEntry{{
			ID:           "fixture-loop",
			LoopSource:   "docs/30-architecture/executable-knowledge.md",
			Observes:     []string{"fixture observation"},
			Verifies:     []string{"fixture verifier"},
			Evidence:     []string{"fixture evidence"},
			ExpiresAfter: "24h",
			FailsWhen:    []string{"fixture failure"},
		}},
		ManualGroups: []manualGroup{{
			ID:           "known",
			Owner:        "test",
			Reason:       "fixture",
			NextArtifact: "fixture",
			Paths:        []string{"docs/30-architecture/known.md"},
		}},
		Assertions: []string{"fixture assertion"},
		Loop: evidenceLoop{
			Observation:   "fixture observation",
			Hypothesis:    "fixture hypothesis",
			Execute:       "fixture execute",
			Evaluate:      "fixture evaluate",
			Retrospective: "fixture retrospective",
		},
	}
}

func writeFixture(t *testing.T, root, path, text string) {
	t.Helper()
	if err := writeText(filepath.Join(root, filepath.FromSlash(path)), text); err != nil {
		t.Fatal(err)
	}
}
