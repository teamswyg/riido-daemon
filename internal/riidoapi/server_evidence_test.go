package riidoapi

import (
	"context"
	"testing"
)

func TestServerReplaysEvidenceCommandIDWithoutDuplicateMutation(t *testing.T) {
	socketPath, _, stop := serveTestAPI(t)
	defer stop()

	client := NewClient(socketPath)
	request := EvidenceRequest{
		TaskID:            "task:test",
		Command:           "cargo check",
		ExitCode:          0,
		Actor:             "daemon",
		Source:            "test",
		Summary:           "build passed",
		ApprovalID:        "approval:test:evidence",
		CommandID:         "command:test:evidence",
		ValidationGate:    "validation:api:cargo-check",
		ProviderRunID:     "provider-run:api:evidence",
		ProviderRunResult: "passed",
	}
	var first EvidenceResponse
	if err := client.Request(context.Background(), "evidence", request, &first); err != nil {
		t.Fatalf("first evidence request failed: %v", err)
	}

	var replayed EvidenceResponse
	if err := client.Request(context.Background(), "evidence", request, &replayed); err != nil {
		t.Fatalf("replayed evidence request failed: %v", err)
	}
	if replayed.Evidence.ID != first.Evidence.ID {
		t.Fatalf("replay should return original evidence: first=%s replay=%s", first.Evidence.ID, replayed.Evidence.ID)
	}
	if replayed.Receipt.ID != first.Receipt.ID {
		t.Fatalf("replay should return original receipt: first=%s replay=%s", first.Receipt.ID, replayed.Receipt.ID)
	}

	var status Status
	if err := client.Request(context.Background(), "status", nil, &status); err != nil {
		t.Fatalf("status request failed: %v", err)
	}
	if status.EvidenceCount != 1 || status.CommandReceiptCount != 1 {
		t.Fatalf("replay should not append state: %#v", status)
	}
}
