package mwsdbridge

import (
	"context"
	"path/filepath"
	"testing"
)

func TestFetchSnapshot(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "mwsd.sock")
	stop := serveFakeMwsd(t, socketPath, fakeSnapshotResponses())
	defer stop()

	snapshot, err := NewClient(socketPath).FetchSnapshot(context.Background())
	if err != nil {
		t.Fatalf("FetchSnapshot returned error: %v", err)
	}
	if snapshot.Status.Root != "/workspace" {
		t.Fatalf("unexpected root: %s", snapshot.Status.Root)
	}
	if snapshot.Graph.Stats.DocumentCount != 23 {
		t.Fatalf("unexpected document count: %d", snapshot.Graph.Stats.DocumentCount)
	}
	if snapshot.Domain.Domain != "macmini-workspace" {
		t.Fatalf("unexpected domain: %s", snapshot.Domain.Domain)
	}
	if snapshot.Harness.NextDirection != "top-down" {
		t.Fatalf("unexpected next direction: %s", snapshot.Harness.NextDirection)
	}
	if got := snapshot.Projects.Repositories[0].Name; got != "riido-daemon" {
		t.Fatalf("unexpected project repository: %s", got)
	}
	if snapshot.Orchestration.RecommendedProvider != "codex" {
		t.Fatalf("unexpected recommended provider: %s", snapshot.Orchestration.RecommendedProvider)
	}
	if len(snapshot.Orchestration.ProviderCandidates) != 3 {
		t.Fatalf("unexpected provider candidate count: %d", len(snapshot.Orchestration.ProviderCandidates))
	}
}
