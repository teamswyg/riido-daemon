package supervisor

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/providerrecovery"
)

func TestPreflightRetryPreservesPreparedWorkspaceAndDoesNotCancelRun(t *testing.T) {
	p := &retryStartProcess{failures: 1, started: make(chan *process.FakeRunning, 2)}
	rt, _ := newRecoveryRuntime(t, "fake", "", p)
	_, reporter, _ := startRecoverySupervisor(t, recoveryRequest(), rt)
	running := awaitRecoverySpawn(t, p)
	p.mu.Lock()
	dirs := append([]string(nil), p.dirs...)
	p.mu.Unlock()
	if len(dirs) != 2 || dirs[0] == "" || dirs[0] != dirs[1] {
		t.Fatalf("retry changed prepared workspace: %v", dirs)
	}
	select {
	case <-running.KillRecv():
		t.Fatal("preflight budget cancellation killed an admitted run")
	case <-time.After(30 * time.Millisecond):
	}
	completeTaskResultProcess(running)
	result := expectTaskResult(t, reporter, "recovery result missing")
	if result.Status != agentbridge.ResultCompleted || result.RecoveryContext == nil || result.RecoveryContext.Attempt != 2 || result.RecoveryContext.State != "recovered" {
		t.Fatalf("result=%+v context=%+v", result, result.RecoveryContext)
	}
}

func TestPreflightRetryBudgetIsFinite(t *testing.T) {
	p := &retryStartProcess{failures: 10, started: make(chan *process.FakeRunning, 2)}
	rt, _ := newRecoveryRuntime(t, "fake", "", p)
	req := recoveryRequest()
	req.RecoveryPolicy = &assignment.ProviderRecoveryPolicy{MaxAttempts: 2, MaxElapsedSeconds: 2}
	_, reporter, _ := startRecoverySupervisor(t, req, rt)
	result := expectTaskResult(t, reporter, "budget never terminated")
	if p.count() != 2 || result.RecoveryContext == nil || result.RecoveryContext.Reason != "budget_exhausted" {
		t.Fatalf("attempts=%d context=%+v", p.count(), result.RecoveryContext)
	}
}

func TestCancellationDuringRecoveryBackoffNeverSpawnsAgain(t *testing.T) {
	p := &retryStartProcess{failures: 10, started: make(chan *process.FakeRunning, 2)}
	rt, _ := newRecoveryRuntime(t, "fake", "", p)
	req := recoveryRequest()
	reporter, source := startBackoffCancellationSupervisor(t, req, rt)
	// A buffered progress event can be observed after all attempts have ended.
	// Hold the first backoff journal and arm the actual source watch instead.
	var backoffCtx context.Context
	select {
	case backoffCtx = <-reporter.backoff:
	case <-time.After(supervisorCancellationTestTimeout):
		t.Fatal("backoff journal boundary missing")
	}
	select {
	case <-source.watched:
	case <-time.After(supervisorCancellationTestTimeout):
		t.Fatal("cancellation watch was not registered")
	}
	if p.count() != 1 {
		t.Fatal("provider retried before the cancellation boundary")
	}
	source.Cancel(req.ID, context.Canceled)
	select {
	case <-backoffCtx.Done():
	case <-time.After(supervisorCancellationTestTimeout):
		t.Fatal("cancellation did not reach the recovery context")
	}
	result := expectTaskResult(t, reporter.reporterProbe, "cancel did not finish recovery")
	if result.Status != agentbridge.ResultCancelled || p.count() != 1 {
		t.Fatalf("cancelled result=%+v attempts=%d", result, p.count())
	}
}

func TestProviderFailureAfterDispatchIsNeverReplayed(t *testing.T) {
	p := &retryStartProcess{started: make(chan *process.FakeRunning, 2)}
	rt, _ := newRecoveryRuntime(t, "fake", "", p)
	_, reporter, _ := startRecoverySupervisor(t, recoveryRequest(), rt)
	running := awaitRecoverySpawn(t, p)
	go running.EmitExit(1, context.DeadlineExceeded)
	result := expectTaskResult(t, reporter, "unknown outcome missing")
	if p.count() != 1 || result.RecoveryContext == nil || result.RecoveryContext.Reason != "unknown_outcome" {
		t.Fatalf("replayed unknown outcome: %d %+v", p.count(), result.RecoveryContext)
	}
}

func TestExplicitFallbackUsesApprovedModelAndNeutralInstructions(t *testing.T) {
	originalProcess := &retryStartProcess{started: make(chan *process.FakeRunning, 2)}
	original, _ := newRecoveryRuntime(t, "claude", hostintegration.ProviderDiagnosticExecutableMissing, originalProcess)
	fallbackProcess := &retryStartProcess{started: make(chan *process.FakeRunning, 2)}
	fallback, adapter := newRecoveryRuntime(t, "codex", "", fallbackProcess)
	req := recoveryRequest()
	req.Provider = "claude"
	req.AllowExperimentalRuntime = true
	req.Prompt = "ORIGINAL_PROVIDER_ONLY private protocol"
	req.AgentInstruction = "Use the task scope"
	req.RecoveryPolicy = &assignment.ProviderRecoveryPolicy{MaxAttempts: 3, MaxSwitches: 1, Fallbacks: []assignment.ProviderRecoveryCandidate{{Provider: "codex", ModelID: "chosen-model"}}}
	actor, reporter, _ := startRecoverySupervisor(t, req, original, fallback)
	var running *process.FakeRunning
	select {
	case running = <-fallbackProcess.started:
	case result := <-reporter.results:
		t.Fatalf("fallback refused: %+v recovery=%+v", result, result.RecoveryContext)
	case <-time.After(4 * time.Second):
		t.Fatal("fallback stuck")
	}
	request := <-adapter.requests
	if originalProcess.count() != 0 || request.Model != "chosen-model" || strings.Contains(request.Prompt, "ORIGINAL_PROVIDER_ONLY") || request.ResumeSessionID != "" {
		t.Fatalf("unsafe handoff: %+v", request)
	}
	ownership := actor.OwnedExecutions()
	if len(ownership["rt:claude"]) != 1 {
		t.Fatalf("assignment lease moved from claim runtime: %+v", ownership)
	}
	completeTaskResultProcess(running)
	result := expectTaskResult(t, reporter, "fallback result missing")
	if result.Status != agentbridge.ResultCompleted || result.RecoveryContext == nil || result.RecoveryContext.SwitchCount != 1 {
		t.Fatalf("result=%+v context=%+v", result, result.RecoveryContext)
	}
}

func TestAuthenticationFailureNeverSwitchesProvider(t *testing.T) {
	p := &retryStartProcess{started: make(chan *process.FakeRunning, 2)}
	original, _ := newRecoveryRuntime(t, "claude", hostintegration.ProviderDiagnosticLoginRequired, p)
	targetProcess := &retryStartProcess{started: make(chan *process.FakeRunning, 2)}
	target, _ := newRecoveryRuntime(t, "codex", "", targetProcess)
	req := recoveryRequest()
	req.Provider = "claude"
	req.RecoveryPolicy = &assignment.ProviderRecoveryPolicy{MaxSwitches: 1, Fallbacks: []assignment.ProviderRecoveryCandidate{{Provider: "codex", ModelID: "chosen-model"}}}
	_, reporter, _ := startRecoverySupervisor(t, req, original, target)
	result := expectTaskResult(t, reporter, "auth failure missing")
	if result.Status != agentbridge.ResultBlocked || targetProcess.count() != 0 || p.count() != 0 {
		t.Fatal("authentication failure switched or spawned provider")
	}
}

func TestFallbackPreservesExactNativeCatalogModelID(t *testing.T) {
	for _, model := range []string{"openai/gpt-test:stable", "auto"} {
		t.Run(model, func(t *testing.T) {
			originalProcess := &retryStartProcess{started: make(chan *process.FakeRunning, 1)}
			original, _ := newRecoveryRuntime(t, "claude", hostintegration.ProviderDiagnosticExecutableMissing, originalProcess)
			targetProcess := &retryStartProcess{started: make(chan *process.FakeRunning, 1)}
			target, adapter := newRecoveryRuntime(t, "openclaw", "", targetProcess, model)
			req := recoveryRequest()
			req.Provider = "claude"
			req.AllowExperimentalRuntime = true
			req.RecoveryPolicy = &assignment.ProviderRecoveryPolicy{MaxSwitches: 1, Fallbacks: []assignment.ProviderRecoveryCandidate{{Provider: "openclaw", ModelID: model}}}
			_, reporter, _ := startRecoverySupervisor(t, req, original, target)
			run := awaitRecoverySpawn(t, targetProcess)
			request := <-adapter.requests
			if request.Model != model {
				t.Fatalf("native catalog id changed: %q != %q", request.Model, model)
			}
			completeTaskResultProcess(run)
			_ = expectTaskResult(t, reporter, "fallback missing")
		})
	}
}

func TestFirstAttemptCompletionIsNotCountedAsRecoveryOrHostVerification(t *testing.T) {
	c := assignment.RecoveryContext{State: "dispatched", Provider: "codex", Reason: "none", Attempt: 1, OutcomeCertainty: "started", Circuit: "closed"}
	task := &runningTask{taskID: "normal", recovery: &c}
	result := terminalRecoveryContext(&Actor{}, task, agentbridge.Result{Status: agentbridge.ResultCompleted})
	if result.RecoveryContext.State != "dispatched" || result.RecoveryContext.OutcomeCertainty != "confirmed" {
		t.Fatal("ordinary completion counted as recovery")
	}
	if resultExecutionContext(task, result).VerificationStatus != "not_configured" {
		t.Fatal("protocol completion counted as host verification")
	}
}

func TestFallbackNativeSessionIsKeptLocallyAndNotPinnedToOriginalAssignment(t *testing.T) {
	actor, status := historyActor(t.TempDir())
	request := historyRequest("fallback", "")
	prepared := prepareAndPinHistory(t, actor, status, request)
	reporter := newReporterProbe()
	actor.cfg.Reporter = reporter
	task := &runningTask{taskID: request.ID, request: request, workspace: prepared.workspace, events: prepared.events, recovery: &assignment.RecoveryContext{SwitchCount: 1, Provider: "claude"}}
	detached := []detachedReport{}
	actor.handleTaskEvent(context.Background(), &taskEventMsg{taskID: request.ID, event: agentbridge.Event{Kind: agentbridge.EventSessionIdentified, SessionID: "foreign-session"}}, map[string]*runningTask{request.ID: task}, &detached)
	if task.sessionBindingErr != nil || task.nativeSessionID != "foreign-session" {
		t.Fatal("handoff lost local native session provenance")
	}
	select {
	case event := <-reporter.events:
		if event.Kind == agentbridge.EventSessionIdentified || event.SessionID != "" {
			t.Fatalf("foreign session reached original assignment reporter: %+v", event)
		}
	default:
	}
}

func TestHostCorrectionKeepsRecoveryScopeAndReleasesHalfOpenAtCompletion(t *testing.T) {
	circuits := providerrecovery.NewCircuits()
	now := time.Now()
	circuits.Failure("scope", "old", now)
	circuits.Failure("scope", "old", now)
	if state, _ := circuits.Acquire("scope", "execution", now.Add(2*time.Second)); state != "half_open" {
		t.Fatal(state)
	}
	c := assignment.RecoveryContext{State: "dispatched", Provider: "claude", Circuit: "half_open", Attempt: 2, SwitchCount: 1, OutcomeCertainty: "started"}
	task := &runningTask{taskID: "execution", recovery: &c, recoveryScope: "scope"}
	applyActivationRecovery(task, &taskActivationMsg{})
	if task.recovery != &c || task.recoveryScope != "scope" || task.recovery.SwitchCount != 1 {
		t.Fatal("host correction discarded recovery/session ownership")
	}
	result := terminalRecoveryContext(&Actor{recovery: circuits}, task, agentbridge.Result{Status: agentbridge.ResultCompleted})
	if result.RecoveryContext.State != "recovered" || result.RecoveryContext.SwitchCount != 1 || result.RecoveryContext.Circuit != "closed" {
		t.Fatal("host correction lost recovery outcome")
	}
	if state, _ := circuits.Acquire("scope", "next", now.Add(2*time.Second)); state != "closed" {
		t.Fatal("host correction leaked half-open owner")
	}
}
