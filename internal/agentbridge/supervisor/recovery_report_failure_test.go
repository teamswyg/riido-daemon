package supervisor

import (
	"context"
	"errors"
	"testing"

	"github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

type recoveryReportFailureReporter struct {
	*reporterProbe
	state string
	err   error
	calls int
}

func (r *recoveryReportFailureReporter) ReportEvent(ctx context.Context, id string, ev agentbridge.Event) error {
	if ev.RecoveryContext != nil && ev.RecoveryContext.State == r.state {
		r.calls++
		return r.err
	}
	return r.reporterProbe.ReportEvent(ctx, id, ev)
}

func TestRecoveryReportFailureDoesNotBlameProvider(t *testing.T) {
	for _, stage := range []string{"checking", "backoff", "switching"} {
		for _, failure := range []struct {
			name   string
			err    error
			reason string
		}{
			{"transport", errors.New("fixture transport failure"), "unknown"},
			{"server", errors.New("fixture server unavailable"), "unknown"},
			{"acknowledgement", errors.New("fixture acknowledgement identity invalid"), "unknown"},
			{"deadline", context.DeadlineExceeded, "deadline_exceeded"},
		} {
			t.Run(stage+"/"+failure.name, func(t *testing.T) {
				req := recoveryRequest()
				p := &retryStartProcess{started: make(chan *process.FakeRunning, 1)}
				unavailable := hostintegration.ProviderDiagnosticCode("")
				if stage == "backoff" {
					p.failures = 1
				}
				if stage == "switching" {
					req.Provider = "claude"
					req.AllowExperimentalRuntime = true
					unavailable = hostintegration.ProviderDiagnosticExecutableMissing
					req.RecoveryPolicy = &assignment.ProviderRecoveryPolicy{MaxAttempts: 3, MaxSwitches: 1, Fallbacks: []assignment.ProviderRecoveryCandidate{{Provider: "codex", ModelID: "chosen-model"}}}
				}
				rt, _ := newRecoveryRuntime(t, string(req.Provider), unavailable, p)
				runtimes := []*runtimeactor.Actor{rt}
				targetProcess := &retryStartProcess{started: make(chan *process.FakeRunning, 1)}
				if stage == "switching" {
					target, _ := newRecoveryRuntime(t, "codex", "", targetProcess)
					runtimes = append(runtimes, target)
				}
				source := controlplane.NewMemorySource()
				source.Enqueue(req)
				reporter := &recoveryReportFailureReporter{reporterProbe: newReporterProbe(), state: stage, err: failure.err}
				startRoutingSupervisor(t, Config{DaemonID: "daemon", Source: source, Reporter: reporter, Runtimes: runtimes, Workdir: workdir.NewFSAdapter(t.TempDir())})
				result := expectTaskResult(t, reporter.reporterProbe, "recovery report failure did not terminate")
				if reporter.calls != 1 {
					t.Fatalf("expected one failed %s report, got %d", stage, reporter.calls)
				}
				if result.Status != agentbridge.ResultFailed {
					t.Fatalf("report failure blamed provider: status=%s", result.Status)
				}
				c := result.RecoveryContext
				if c == nil || c.State != "unavailable" || c.Reason != failure.reason || c.OutcomeCertainty != "not_started" {
					t.Fatalf("incorrect recovery observation: %+v", c)
				}
				if result.ExecutionContext == nil || result.ExecutionContext.DiagnosticCode != "unknown" || result.ExecutionContext.NextAction != "review_result" {
					t.Fatalf("incorrect execution diagnostic: %+v", result.ExecutionContext)
				}
				_, payload := terminalResultDraft(result)
				if payload["category"] != "provider_result_failed" {
					t.Fatalf("incorrect terminal category: %v", payload["category"])
				}
				wantCalls := 0
				if stage == "backoff" {
					wantCalls = 1
				}
				if p.count() != wantCalls || targetProcess.count() != 0 {
					t.Fatal("report failure allowed provider retry or handoff")
				}
			})
		}
	}
}
