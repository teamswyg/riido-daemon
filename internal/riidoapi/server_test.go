package riidoapi

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func TestServerExposesTaskDB(t *testing.T) {
	socketPath, taskDBPath, stop := serveTestAPI(t)
	defer stop()

	client := NewClient(socketPath)
	var status Status
	if err := client.Request(context.Background(), "status", nil, &status); err != nil {
		t.Fatalf("status request failed: %v", err)
	}
	if status.SchemaVersion != StatusSchemaVersion {
		t.Fatalf("unexpected status schema: %s", status.SchemaVersion)
	}
	if status.Transport != string(LocalTransportUnixSocket) {
		t.Fatalf("unexpected API transport: %s", status.Transport)
	}
	if status.TaskCount != 1 {
		t.Fatalf("unexpected task count: %d", status.TaskCount)
	}
	if status.EvidenceCount != 0 {
		t.Fatalf("unexpected evidence count: %d", status.EvidenceCount)
	}
	if status.CommandReceiptCount != 0 {
		t.Fatalf("unexpected command receipt count: %d", status.CommandReceiptCount)
	}
	if status.TaskDBPath != taskDBPath {
		t.Fatalf("unexpected task DB path: %s", status.TaskDBPath)
	}

	var db taskdb.TaskDB
	if err := client.Request(context.Background(), "tasks", nil, &db); err != nil {
		t.Fatalf("tasks request failed: %v", err)
	}
	if db.SchemaVersion != taskdb.TaskDBSchemaVersion {
		t.Fatalf("unexpected task DB schema: %s", db.SchemaVersion)
	}
	if db.Tasks[0].ID != "task:test" {
		t.Fatalf("unexpected task: %#v", db.Tasks[0])
	}
	if db.Tasks[0].RecommendedProvider != "codex" {
		t.Fatalf("unexpected task recommended provider: %s", db.Tasks[0].RecommendedProvider)
	}
	if !db.Tasks[0].RequiresHumanApproval {
		t.Fatalf("task queue should expose human approval gate")
	}
}

func TestServerAppliesTransition(t *testing.T) {
	socketPath, _, stop := serveTestAPI(t)
	defer stop()

	client := NewClient(socketPath)
	var response TransitionResponse
	err := client.Request(context.Background(), "transition", TransitionRequest{
		TaskID:     "task:test",
		ToState:    string(task.StateQueued),
		EventType:  string(ir.EventTaskQueued),
		Actor:      "human",
		Source:     "test",
		Reason:     "ready",
		ApprovalID: "approval:test:transition",
		CommandID:  "command:test:transition",
	}, &response)
	if err != nil {
		t.Fatalf("transition request failed: %v", err)
	}
	if response.Task.State != task.StateQueued {
		t.Fatalf("unexpected task state: %s", response.Task.State)
	}
	if response.Transition.EventType != ir.EventTaskQueued {
		t.Fatalf("unexpected transition: %#v", response.Transition)
	}
	if response.Receipt.ID == "" || response.Receipt.ID != response.Transition.CommandReceiptID {
		t.Fatalf("transition should return command receipt: %#v", response)
	}
}

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
	if response.Evidence.ProviderRunID != "provider-run:api:evidence" || response.Evidence.ProviderRunResult != "passed" {
		t.Fatalf("unexpected provider run evidence: %#v", response.Evidence)
	}
	if response.Receipt.ID == "" || response.Receipt.ID != response.Evidence.CommandReceiptID {
		t.Fatalf("evidence should return command receipt: %#v", response)
	}
}

func TestServerValidatesTaskAndTransitionsFromValidating(t *testing.T) {
	socketPath, _, stop := serveTestAPIWithState(t, task.StateValidating)
	defer stop()

	client := NewClient(socketPath)
	var response ValidateResponse
	err := client.Request(context.Background(), "validate", ValidateRequest{
		TaskID:         "task:test",
		Command:        "printf ok",
		Workdir:        t.TempDir(),
		TimeoutSeconds: 5,
		Actor:          "daemon",
		Source:         "test",
		ApprovalID:     "approval:test:validate",
		CommandID:      "command:test:validate",
		Provider:       "codex",
		DecisionLLM:    "codex",
	}, &response)
	if err != nil {
		t.Fatalf("validate request failed: %v", err)
	}
	if response.Validation.Result != "passed" {
		t.Fatalf("unexpected validation result: %#v", response.Validation)
	}
	if response.Evidence.Result != "passed" || response.Evidence.CommandReceiptID != response.Receipt.ID {
		t.Fatalf("unexpected evidence/receipt pair: %#v", response)
	}
	if response.Task.State != task.StatePatchReady {
		t.Fatalf("expected task to transition to PatchReady, got %s", response.Task.State)
	}
	if response.Transition == nil || response.Transition.EventType != ir.EventValidationPassed {
		t.Fatalf("expected validation passed transition: %#v", response.Transition)
	}
	if response.TransitionReceipt == nil || response.TransitionReceipt.CommandID != "command:test:validate:transition" {
		t.Fatalf("expected transition receipt pair: %#v", response.TransitionReceipt)
	}
}

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

func TestServerRejectsIllegalTransition(t *testing.T) {
	socketPath, _, stop := serveTestAPI(t)
	defer stop()

	client := NewClient(socketPath)
	var response TransitionResponse
	err := client.Request(context.Background(), "transition", TransitionRequest{
		TaskID:    "task:test",
		ToState:   string(task.StateRunning),
		EventType: string(ir.EventRunStarted),
	}, &response)
	if err == nil {
		t.Fatal("expected illegal transition error")
	}
}

func serveTestAPI(t *testing.T) (string, string, func()) {
	t.Helper()
	return serveTestAPIWithState(t, task.StateCreated)
}

func serveTestAPIWithState(t *testing.T, state task.TaskState) (string, string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "rapi-")
	if err != nil {
		t.Fatalf("MkdirTemp failed: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})
	socketPath := filepath.Join(dir, "riido.sock")
	taskDBPath := filepath.Join(dir, "task-db.json")
	db := taskdb.EmptyTaskDB()
	db.UpdatedAt = "2026-05-20T08:00:00Z"
	db.RecommendedProvider = "codex"
	db.RecommendedDecisionLLM = "codex"
	db.DecisionGate = "human-approval-required"
	db.ProviderCandidates = []taskdb.ProviderCandidate{
		{ID: "codex", SourceWorkflow: "provider-selection", Available: true, ApprovalRequired: true},
	}
	db.Tasks = []taskdb.TaskRecord{
		{
			ID:                     "task:test",
			ProjectID:              "macmini-workspace",
			State:                  state,
			SourceDocumentID:       "mws.test",
			SourceDocumentPath:     "docs/TEST.md",
			Title:                  "테스트",
			Owner:                  "local",
			SourceStatus:           "seed",
			RecommendedProvider:    "codex",
			RecommendedDecisionLLM: "codex",
			RequiresHumanApproval:  true,
			HarnessNextDirection:   "top-down",
			CreatedAt:              "2026-05-20T08:00:00Z",
			UpdatedAt:              "2026-05-20T08:00:00Z",
			TransitionCount:        1,
		},
	}
	db.Transitions = []taskdb.TaskTransitionRecord{
		{
			ID:         "transition:task:test:created",
			TaskID:     "task:test",
			ToState:    task.StateCreated,
			EventType:  ir.EventTaskCreated,
			Actor:      "riido",
			Source:     "test",
			RecordedAt: "2026-05-20T08:00:00Z",
		},
	}
	if err := taskdb.SaveTaskDB(taskDBPath, db); err != nil {
		t.Fatalf("SaveTaskDB failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	errs := make(chan error, 1)
	go func() {
		errs <- NewServer(Config{SocketPath: socketPath, TaskDBPath: taskDBPath}).Serve(ctx)
	}()
	waitForAPI(t, socketPath)
	return socketPath, taskDBPath, func() {
		cancel()
		select {
		case err := <-errs:
			if err != nil {
				t.Fatalf("Serve returned error: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for API server shutdown")
		}
	}
}

func waitForAPI(t *testing.T, socketPath string) {
	t.Helper()
	client := NewClient(socketPath)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		var status Status
		if err := client.Request(context.Background(), "status", nil, &status); err == nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("riido API did not become ready at %s", socketPath)
}

func sameStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range got {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}
