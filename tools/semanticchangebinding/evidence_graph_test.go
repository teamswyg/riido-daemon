package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestEvidenceGraphRecordsVerifiedDecision(t *testing.T) {
	evidence := runBusinessClaimEvidence(t, businessClaimPeerFiles())
	entry := requireGraphEntry(t, evidence, "taskdb-lease-fencing-claim")
	if entry.Decision != "verified" || entry.NextLoop == "" {
		t.Fatalf("unexpected verified graph entry: %#v", entry)
	}
	if len(entry.Observation) == 0 || len(entry.Change) == 0 {
		t.Fatalf("expected observation and change paths: %#v", entry)
	}
}

func TestEvidenceGraphRecordsMissingPeerDecision(t *testing.T) {
	evidence := runBusinessClaimEvidence(t, []string{
		"internal/agentbridge/controlplane/taskdbplane/runtime_lease_require.go",
	})
	entry := requireGraphEntry(t, evidence, "taskdb-lease-fencing-claim")
	if entry.Decision != "failed_missing_semantic_peers" {
		t.Fatalf("unexpected failed graph entry: %#v", entry)
	}
}

func runBusinessClaimEvidence(t *testing.T, changed []string) Evidence {
	t.Helper()
	repo := fixtureRepo(t)
	out := filepath.Join(t.TempDir(), "semantic.json")
	_ = run(context.Background(), options{Repo: repo, Manifest: manifestPath(), ChangedFiles: changed, EvidenceOut: out})
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var evidence Evidence
	if err := json.Unmarshal(data, &evidence); err != nil {
		t.Fatal(err)
	}
	return evidence
}

func requireGraphEntry(t *testing.T, evidence Evidence, id string) evidenceGraphEntry {
	t.Helper()
	for _, entry := range evidence.EvidenceGraph {
		if entry.BindingID == id {
			return entry
		}
	}
	t.Fatalf("missing graph entry %s in %#v", id, evidence.EvidenceGraph)
	return evidenceGraphEntry{}
}
