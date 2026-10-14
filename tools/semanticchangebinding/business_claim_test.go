package main

import (
	"context"
	"strings"
	"testing"
)

func TestBusinessClaimCodeOnlyChangeFails(t *testing.T) {
	repo := fixtureRepo(t)
	err := run(context.Background(), options{
		Repo: repo, Manifest: manifestPath(),
		ChangedFiles: []string{
			"internal/agentbridge/controlplane/taskdbplane/runtime_lease_require.go",
		},
	})
	if err == nil || !strings.Contains(err.Error(), "taskdb-lease-fencing-claim") {
		t.Fatalf("expected business claim peer failure, got %v", err)
	}
}

func TestBusinessClaimFullPeerChangePasses(t *testing.T) {
	repo := fixtureRepo(t)
	err := run(context.Background(), options{
		Repo: repo, Manifest: manifestPath(),
		ChangedFiles: businessClaimPeerFiles(),
	})
	if err != nil {
		t.Fatalf("expected business claim peers to pass, got %v", err)
	}
}

func businessClaimPeerFiles() []string {
	return []string{
		"docs/20-domain/runtime-scheduling/invariants/local-daemon-contract.riido.json",
		"docs/20-domain/runtime-scheduling/invariants/local-daemon-contract.md",
		"internal/agentbridge/controlplane/taskdbplane/task_request_from_record.go",
		"internal/agentbridge/controlplane/taskdbplane/runtime_lease_require.go",
		"internal/agentbridge/controlplane/taskdbplane/task_claim_lease_metadata_test.go",
		"internal/agentbridge/controlplane/taskdbplane/runtime_lease_start_reject_test.go",
		".github/workflows/local-daemon-contract-evidence.yml",
	}
}
