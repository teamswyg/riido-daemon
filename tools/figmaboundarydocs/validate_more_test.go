package main

import (
	"strings"
	"testing"
)

func TestBuildEvidenceValidateAndProblemMessages(t *testing.T) {
	m := manifest{
		ID:            "figma-boundary",
		SchemaVersion: "riido-figma-ai-agent-daemon-boundary.v1",
		RiidoTask:     "RIID-1",
		HumanDoc:      "doc",
		BoundaryPolicy: boundaryPolicy{
			Summary: "summary", TopDown: "top", BottomUp: "bottom",
		},
		Entries: []boundaryEntry{{
			NodeID: "n1", Name: "Screen", DaemonScope: "runtime",
			UpstreamOwner: []string{"figma"}, DaemonConsumedFacts: []string{},
			ClientOwnedFacts: []string{"copy"},
		}},
	}
	if problems := validateManifest(m); len(problems) != 0 {
		t.Fatalf("valid manifest problems = %+v", problems)
	}
	ev := buildEvidence(m, []problem{{Message: "bad"}})
	if ev.Status != "failed" || ev.EntryCount != 1 || len(ev.GeneratedDocs) != 7 {
		t.Fatalf("evidence = %+v", ev)
	}
	if !strings.Contains(problemError([]problem{{Message: "bad"}}).Error(), "bad") {
		t.Fatal("problem error should include message")
	}
}

func TestValidateEntryAndFixtureFallback(t *testing.T) {
	if problems := validateEntry(boundaryEntry{}); len(problems) != 1 {
		t.Fatalf("empty entry problems = %+v", problems)
	}
	m := manifest{Entries: []boundaryEntry{{Name: "Other", ClientOwnedFacts: []string{"copy"}}}}
	if entry := findFixtureEntry(m); entry.NodeID != "" {
		t.Fatalf("unexpected fixture entry = %+v", entry)
	}
}
