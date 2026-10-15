package supervisor

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/execution"
	"github.com/teamswyg/riido-daemon/internal/process"
)

type readyRetryReporter struct {
	*reporterProbe
	mu        sync.Mutex
	failures  int
	calls     int
	attempted chan int
	release   chan struct{}
}

func (r *readyRetryReporter) StartTask(ctx context.Context, id string) error {
	r.mu.Lock()
	r.calls++
	attempt := r.calls
	r.mu.Unlock()
	select {
	case r.attempted <- attempt:
	default:
	}
	if r.release != nil {
		select {
		case <-r.release:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	if attempt <= r.failures {
		return transientReadyFixtureError{}
	}
	return r.reporterProbe.StartTask(ctx, id)
}

type transientReadyFixtureError struct{}

func (transientReadyFixtureError) Error() string        { return "temporary Ready transport failure" }
func (transientReadyFixtureError) RetryableReady() bool { return true }
func newReadyRetryReporter(failures int) *readyRetryReporter {
	return &readyRetryReporter{reporterProbe: newReporterProbe(), failures: failures, attempted: make(chan int, 10)}
}
func (r *readyRetryReporter) count() int { r.mu.Lock(); defer r.mu.Unlock(); return r.calls }

func TestPendingReadyRetainsLeaseAndRetriesBeforeSingleDispatch(t *testing.T) {
	source := controlplane.NewMemorySource()
	req := dirtyWorkdirDriftRequest()
	source.Enqueue(req)
	reporter := newReadyRetryReporter(1)
	reporter.release = make(chan struct{})
	p := &retryStartProcess{started: make(chan *process.FakeRunning, 2)}
	rt, _ := newRecoveryRuntime(t, string(req.Provider), "", p)
	actor := startRoutingSupervisor(t, Config{DaemonID: "daemon", Runtime: rt, Source: source, Reporter: reporter, HeartbeatEvery: 10 * time.Millisecond})
	select {
	case <-reporter.attempted:
	case <-time.After(time.Second):
		t.Fatal("Ready never attempted")
	}
	deadline := time.Now().Add(time.Second)
	for len(actor.OwnedExecutions()["rt:"+string(req.Provider)]) != 1 {
		if time.Now().After(deadline) {
			t.Fatal("pending Ready claim lost heartbeat ownership")
		}
		time.Sleep(time.Millisecond)
	}
	if p.count() != 0 {
		t.Fatal("provider started before Ready ACK")
	}
	close(reporter.release)
	run := awaitRecoverySpawn(t, p)
	if reporter.count() != 2 || p.count() != 1 {
		t.Fatalf("Ready calls=%d dispatches=%d", reporter.count(), p.count())
	}
	completeTaskResultProcess(run)
	if result := expectTaskResult(t, reporter.reporterProbe, "result missing"); result.Status != agentbridge.ResultCompleted {
		t.Fatal(result.Status)
	}
}

func TestPendingReadyRetryBudgetExhaustionNeverDispatches(t *testing.T) {
	source := controlplane.NewMemorySource()
	req := dirtyWorkdirDriftRequest()
	source.Enqueue(req)
	reporter := newReadyRetryReporter(10)
	p := &retryStartProcess{started: make(chan *process.FakeRunning, 1)}
	rt, _ := newRecoveryRuntime(t, string(req.Provider), "", p)
	startRoutingSupervisor(t, Config{DaemonID: "daemon", Runtime: rt, Source: source, Reporter: reporter})
	result := expectTaskResult(t, reporter.reporterProbe, "Ready exhaustion lost pending claim")
	if reporter.count() != 3 || p.count() != 0 || result.Status != agentbridge.ResultFailed {
		t.Fatalf("calls=%d starts=%d result=%+v", reporter.count(), p.count(), result)
	}
}

func TestPendingReadyCancellationPreventsAdmission(t *testing.T) {
	source := controlplane.NewMemorySource()
	req := dirtyWorkdirDriftRequest()
	source.Enqueue(req)
	reporter := newReadyRetryReporter(0)
	reporter.release = make(chan struct{})
	p := &retryStartProcess{started: make(chan *process.FakeRunning, 1)}
	rt, _ := newRecoveryRuntime(t, string(req.Provider), "", p)
	startRoutingSupervisor(t, Config{DaemonID: "daemon", Runtime: rt, Source: source, Reporter: reporter})
	select {
	case <-reporter.attempted:
	case <-time.After(time.Second):
		t.Fatal("Ready missing")
	}
	// Wait until the cancellation watch has been registered, rather than racing its setup.
	time.Sleep(20 * time.Millisecond)
	source.Cancel(req.ID, context.Canceled)
	result := expectTaskResult(t, reporter.reporterProbe, "pending Ready did not cancel")
	if result.Status != agentbridge.ResultCancelled || p.count() != 0 {
		t.Fatalf("result=%+v starts=%d", result, p.count())
	}
}

func TestPendingReadyPermanentFailureDoesNotRetry(t *testing.T) {
	reporter := newStartFailReporter()
	a := &Actor{cfg: Config{Reporter: reporter}}
	err := a.acknowledgeReady(context.Background(), "execution")
	if err == nil || errors.Is(err, context.Canceled) {
		t.Fatalf("permanent Ready error=%v", err)
	}
	var failure *readyAcknowledgementFailure
	if !errors.As(err, &failure) || !failure.blocked {
		t.Fatalf("permanent error did not block: %v", err)
	}
	if len(reporter.attempted) != 1 {
		t.Fatalf("attempts=%d", len(reporter.attempted))
	}
}

func TestPendingReadyStopAndLateAcknowledgementNeverDispatch(t *testing.T) {
	source := controlplane.NewMemorySource()
	req := dirtyWorkdirDriftRequest()
	source.Enqueue(req)
	reporter := newReadyRetryReporter(0)
	reporter.release = make(chan struct{})
	p := &retryStartProcess{started: make(chan *process.FakeRunning, 1)}
	rt, _ := newRecoveryRuntime(t, string(req.Provider), "", p)
	actor := startRoutingSupervisor(t, Config{DaemonID: "daemon", Runtime: rt, Source: source, Reporter: reporter})
	select {
	case <-reporter.attempted:
	case <-time.After(time.Second):
		t.Fatal("Ready missing")
	}
	stopCtx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	if err := actor.Stop(stopCtx); err != nil {
		t.Fatalf("Ready held Stop: %v", err)
	}
	close(reporter.release)
	if result := expectTaskResult(t, reporter.reporterProbe, "stop result missing"); result.Status != agentbridge.ResultCancelled {
		t.Fatal(result.Status)
	}
	select {
	case <-p.started:
		t.Fatal("late Ready admitted after Stop")
	case <-time.After(30 * time.Millisecond):
	}
}

func TestObsoleteClaimActivationCannotMutateNewPendingClaim(t *testing.T) {
	oldCtx, cancelOld := context.WithCancel(context.Background())
	cancelOld()
	newCtx := t.Context()
	replacement := &runningTask{taskID: "same", ctx: newCtx}
	actor := &Actor{cfg: Config{Reporter: newReporterProbe()}}
	pending := map[string]*runningTask{"same": replacement}
	actor.handleTaskActivation(context.Background(), &taskActivationMsg{taskID: "same", claimContext: oldCtx, err: context.Canceled}, pending)
	if pending["same"] != replacement || replacement.terminalResult != nil {
		t.Fatal("late old completion terminated the new pending claim")
	}
}

func TestPendingReadyRuntimeDriftAndLateACKNeverAdmit(t *testing.T) {
	source := controlplane.NewMemorySource()
	req := dirtyWorkdirDriftRequest()
	req.Provider = "codex"
	source.Enqueue(req)
	reporter := newReadyRetryReporter(0)
	reporter.release = make(chan struct{})
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	adapter := newMutableDetectAdapter("codex", "1.0.0")
	rt := startRuntimeWithAdapter(t, fake, "rt-codex", adapter)
	startRoutingSupervisor(t, Config{DaemonID: "daemon", Runtime: rt, Source: source, Reporter: reporter, HeartbeatEvery: 10 * time.Millisecond})
	select {
	case <-reporter.attempted:
	case <-time.After(time.Second):
		t.Fatal("Ready missing")
	}
	adapter.setVersion("2.0.0")
	result := expectTaskResult(t, reporter.reporterProbe, "pending Ready drift did not terminate")
	close(reporter.release)
	if result.Status != agentbridge.ResultBlocked {
		t.Fatalf("drift status=%s", result.Status)
	}
	assertProcessDoesNotStart(t, running, "late Ready admitted after runtime drift")
}

func TestHostCorrectionActivationKeepsOriginalClaimContextFence(t *testing.T) {
	oldCtx, cancelOld := context.WithCancel(context.Background())
	cancelOld()
	p := &retryStartProcess{started: make(chan *process.FakeRunning, 1)}
	rt, _ := newRecoveryRuntime(t, "fake", "", p)
	actor := &Actor{cfg: Config{Reporter: newReporterProbe()}, mailbox: make(chan envelope, 1), stoppedCh: make(chan struct{})}
	request := &bridge.TaskRequest{ID: "same", Provider: "fake", Prompt: "fix"}
	oldTask := &runningTask{taskID: request.ID, ctx: oldCtx, runtime: rt, request: request, execution: &execution.Prepared{Binding: execution.Binding{ContinuationLimit: 1}}}
	actor.continueExecution(context.Background(), oldTask, agentbridge.Result{SessionID: "native"}, execution.Evidence{ErrorKind: "unchanged_output"})
	select {
	case message := <-actor.mailbox:
		activation := message.taskActivation
		if activation == nil || activation.claimContext != oldCtx {
			t.Fatal("host correction lost original claim generation")
		}
		replacement := &runningTask{taskID: request.ID, ctx: t.Context()}
		pending := map[string]*runningTask{request.ID: replacement}
		actor.handleTaskActivation(context.Background(), activation, pending)
		if pending[request.ID] != replacement || replacement.terminalResult != nil {
			t.Fatal("late host correction mutated replacement claim")
		}
	case <-time.After(time.Second):
		t.Fatal("correction activation missing")
	}
	if p.count() != 0 {
		t.Fatal("cancelled correction started provider")
	}
}
