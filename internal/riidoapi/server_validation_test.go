package riidoapi

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
)

func TestServerEvaluatesReviewDemoMode(t *testing.T) {
	socketPath, _, stop := serveTestAPI(t)
	defer stop()

	client := NewClient(socketPath)
	var response ReviewDemoResponse
	err := client.Request(context.Background(), "review-demo", ReviewDemoRequest{
		DistributionChannel:      "mac-app-store",
		ReviewDemoConsentGranted: true,
	}, &response)
	if err != nil {
		t.Fatalf("review-demo request failed: %v", err)
	}
	if response.SchemaVersion != ReviewDemoSchemaVersion {
		t.Fatalf("unexpected review-demo schema: %s", response.SchemaVersion)
	}
	if !response.Enabled {
		t.Fatal("review demo mode should be enabled")
	}
	if response.ProviderExecutionAllowed {
		t.Fatal("review demo mode must not allow provider execution")
	}
	if response.TelemetrySyncAllowed {
		t.Fatal("review demo mode must not allow telemetry sync")
	}
	if !response.LocalOnly {
		t.Fatal("review demo mode should be reported as local-only")
	}
	if response.ProviderStatusMode != "synthetic-preview" {
		t.Fatalf("unexpected provider status mode: %s", response.ProviderStatusMode)
	}
	want := []string{"onboarding", "provider-status", "workspace-grant", "background-consent", "privacy-settings", "local-status"}
	if !sameStrings(response.Surfaces, want) {
		t.Fatalf("unexpected surfaces: got %#v want %#v", response.Surfaces, want)
	}
}

func TestServerReviewDemoRequiresConsentForStoreManagedChannel(t *testing.T) {
	socketPath, _, stop := serveTestAPI(t)
	defer stop()

	client := NewClient(socketPath)
	var response ReviewDemoResponse
	err := client.Request(context.Background(), "review-demo", ReviewDemoRequest{
		DistributionChannel: "msix-store",
	}, &response)
	if err == nil {
		t.Fatal("expected review-demo request without consent to fail")
	}
}

func TestServerReviewDemoIgnoresNonStoreManagedChannel(t *testing.T) {
	socketPath, _, stop := serveTestAPI(t)
	defer stop()

	client := NewClient(socketPath)
	var response ReviewDemoResponse
	err := client.Request(context.Background(), "review-demo", ReviewDemoRequest{
		DistributionChannel: "developer-id",
	}, &response)
	if err != nil {
		t.Fatalf("review-demo non-store request failed: %v", err)
	}
	if response.Enabled {
		t.Fatal("non-store channel should not enable review demo mode")
	}
	if response.ProviderStatusMode != "real-status" {
		t.Fatalf("unexpected provider status mode: %s", response.ProviderStatusMode)
	}
	if len(response.Surfaces) != 0 {
		t.Fatalf("non-store channel should not expose synthetic surfaces: %#v", response.Surfaces)
	}
}

func TestServerValidateUsesCallerContext(t *testing.T) {
	_, taskDBPath, stop := serveTestAPIWithState(t, task.StateValidating)
	defer stop()

	params, err := json.Marshal(ValidateRequest{
		TaskID:         "task:test",
		Command:        "sleep 2",
		TimeoutSeconds: 30,
		ApprovalID:     "approval:test:validate-cancel",
		CommandID:      "command:test:validate-cancel",
	})
	if err != nil {
		t.Fatalf("marshal validate params: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	started := time.Now()
	response, err := NewServer(Config{TaskDBPath: taskDBPath}).validateTask(ctx, params)
	if err != nil {
		t.Fatalf("validateTask returned error: %v", err)
	}
	if elapsed := time.Since(started); elapsed > 500*time.Millisecond {
		t.Fatalf("validateTask ignored caller context: elapsed=%s", elapsed)
	}
	if response.Validation.Result != "failed" || response.Validation.ExitCode == 0 {
		t.Fatalf("canceled validation should be recorded as failed: %#v", response.Validation)
	}
}

func TestServerValidateRequiresApprovalBeforeExecution(t *testing.T) {
	socketPath, _, stop := serveTestAPIWithState(t, task.StateValidating)
	defer stop()

	marker := filepath.Join(t.TempDir(), "should-not-exist")
	client := NewClient(socketPath)
	var response ValidateResponse
	err := client.Request(context.Background(), "validate", ValidateRequest{
		TaskID:         "task:test",
		Command:        "touch " + marker,
		Workdir:        t.TempDir(),
		TimeoutSeconds: 5,
		CommandID:      "command:test:validate",
		Provider:       "codex",
		DecisionLLM:    "codex",
	}, &response)
	if err == nil {
		t.Fatal("expected missing approval_id to be rejected")
	}
	if _, statErr := os.Stat(marker); !os.IsNotExist(statErr) {
		t.Fatalf("validation command should not execute without approval, statErr=%v", statErr)
	}
}

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

func TestServerRejectsMissingApproval(t *testing.T) {
	socketPath, _, stop := serveTestAPI(t)
	defer stop()

	client := NewClient(socketPath)
	var response TransitionResponse
	err := client.Request(context.Background(), "transition", TransitionRequest{
		TaskID:    "task:test",
		ToState:   string(task.StateQueued),
		EventType: string(ir.EventTaskQueued),
		Actor:     "human",
		Source:    "test",
	}, &response)
	if err == nil {
		t.Fatal("expected missing approval_id to be rejected")
	}
}
