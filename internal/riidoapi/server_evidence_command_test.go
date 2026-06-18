package riidoapi

import (
	"context"
	"testing"
)

func TestServerAddsEvidence(t *testing.T) {
	socketPath, _, stop := serveTestAPI(t)
	defer stop()

	client := NewClient(socketPath)
	var response EvidenceResponse
	err := client.Request(context.Background(), "evidence", EvidenceRequest{
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
	}, &response)
	if err != nil {
		t.Fatalf("evidence request failed: %v", err)
	}
	if response.Evidence.Result != "passed" {
		t.Fatalf("unexpected evidence result: %s", response.Evidence.Result)
	}
	if response.Evidence.DocumentID != "mws.test" {
		t.Fatalf("unexpected evidence document id: %s", response.Evidence.DocumentID)
	}
	if response.Task.EvidenceCount != 1 {
		t.Fatalf("unexpected task evidence count: %d", response.Task.EvidenceCount)
	}
	if response.Evidence.ValidationGate != "validation:api:cargo-check" {
		t.Fatalf("unexpected validation gate: %s", response.Evidence.ValidationGate)
	}
	if response.Evidence.ProviderRunID != "provider-run:api:evidence" ||
		response.Evidence.ProviderRunResult != "passed" {
		t.Fatalf("unexpected provider run evidence: %#v", response.Evidence)
	}
	if response.Receipt.ID == "" || response.Receipt.ID != response.Evidence.CommandReceiptID {
		t.Fatalf("evidence should return command receipt: %#v", response)
	}
}
