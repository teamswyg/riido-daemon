package supervisor

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func (a *Actor) run(ctx context.Context, runtimes []*runtimeactor.Actor) {
	defer close(a.stoppedCh)
	poll := time.NewTimer(a.cfg.PollEvery)
	defer stopTimer(poll)
	heartbeat := time.NewTicker(a.cfg.HeartbeatEvery)
	defer heartbeat.Stop()

	inFlight := map[string]*runningTask{}
	var stopErr error

	for {
		select {
		case <-ctx.Done():
			stopErr = ctx.Err()
			shutdownCtx, cancel := supervisorShutdownContext(lifecycle.FromContext(ctx).ShutdownLevel())
			a.shutdown(shutdownCtx, runtimes, inFlight)
			cancel()
			a.stopErrCh <- stopErr
			return

		case level := <-a.stopReqCh:
			shutdownCtx, cancel := supervisorShutdownContext(a.drainStopLevel(level))
			a.shutdown(shutdownCtx, runtimes, inFlight)
			cancel()
			a.stopErrCh <- nil
			return

		case <-poll.C:
			claimed := a.claimAvailable(ctx, runtimes, inFlight)
			resetTimer(poll, a.nextPollInterval(claimed, len(inFlight)))

		case <-heartbeat.C:
			for _, rt := range runtimes {
				hb, err := rt.HeartbeatPayload(ctx)
				if err != nil {
					continue
				}
				_ = a.cfg.Source.Heartbeat(ctx, controlplane.RuntimeHeartbeat{
					RuntimeID:      hb.RuntimeID,
					UptimeSeconds:  hb.UptimeSeconds,
					DeviceName:     hb.DeviceName,
					SlotLimit:      hb.SlotLimit,
					SlotsInUse:     hb.SlotsInUse,
					RunningTaskIDs: hb.RunningTaskIDs,
				})
			}

		case msg := <-a.mailbox:
			switch {
			case msg.taskActivation != nil:
				task := inFlight[msg.taskActivation.taskID]
				if task == nil {
					continue
				}
				reportCtx := controlplane.ContextWithTaskReport(ctx, task.report)
				if msg.taskActivation.err != nil {
					if task.cancel != nil {
						task.cancel()
						task.cancel = nil
					}
					res := resultForActivationError(msg.taskActivation.err)
					if msg.taskActivation.prepared != nil {
						task.workspace = msg.taskActivation.prepared.workspace
						task.events = msg.taskActivation.prepared.events
						res = a.recordTerminalResult(ctx, task, res)
					}
					_ = a.cfg.Reporter.CompleteTask(reportCtx, task.taskID, res)
					delete(inFlight, task.taskID)
					resetTimer(poll, a.cfg.PollEvery)
					continue
				}
				if msg.taskActivation.handle == nil {
					_ = a.cfg.Reporter.CompleteTask(reportCtx, task.taskID, agentbridge.Result{
						Status: agentbridge.ResultFailed,
						Error:  "supervisor: runtime submit returned no session handle",
					})
					if task.cancel != nil {
						task.cancel()
						task.cancel = nil
					}
					delete(inFlight, task.taskID)
					resetTimer(poll, a.cfg.PollEvery)
					continue
				}
				task.handle = msg.taskActivation.handle
				if msg.taskActivation.prepared != nil {
					task.workspace = msg.taskActivation.prepared.workspace
					task.events = msg.taskActivation.prepared.events
				}
				if task.cancelCause != nil || task.ctx.Err() != nil {
					a.cancelActivatedTask(ctx, task)
					go a.forwardSession(task.taskID, task.handle.Events(), task.handle.Result())
					continue
				}
				_ = a.cfg.Reporter.ReportEvent(reportCtx, task.taskID, agentbridge.Event{
					Kind:  agentbridge.EventLifecycle,
					Phase: agentbridge.StateRunning,
				})
				go a.forwardSession(task.taskID, task.handle.Events(), task.handle.Result())
			case msg.taskEvent != nil:
				reportCtx := ctx
				if task := inFlight[msg.taskEvent.taskID]; task != nil {
					reportCtx = controlplane.ContextWithTaskReport(ctx, task.report)
					a.appendProviderEvent(ctx, msg.taskEvent.taskID, task.events, msg.taskEvent.event)
				}
				_ = a.cfg.Reporter.ReportEvent(reportCtx, msg.taskEvent.taskID, msg.taskEvent.event)
			case msg.taskResult != nil:
				running := inFlight[msg.taskResult.taskID]
				reportCtx := ctx
				res := msg.taskResult.result
				if running != nil {
					reportCtx = controlplane.ContextWithTaskReport(ctx, running.report)
					res = a.recordTerminalResult(ctx, running, msg.taskResult.result)
					if running.cancel != nil {
						running.cancel()
						running.cancel = nil
					}
				}
				_ = a.cfg.Reporter.CompleteTask(reportCtx, msg.taskResult.taskID, res)
				delete(inFlight, msg.taskResult.taskID)
				resetTimer(poll, a.cfg.PollEvery)
			case msg.cancel != nil:
				if task := inFlight[msg.cancel.taskID]; task != nil {
					task.cancelCause = cancellationCause(msg.cancel.cause)
					if task.cancel != nil {
						task.cancel()
						task.cancel = nil
					}
					if task.handle != nil {
						_ = task.runtime.Cancel(ctx, task.taskID, task.cancelCause.Error())
					}
				}
			}
		}
	}
}

func (a *Actor) cancelActivatedTask(ctx context.Context, task *runningTask) {
	if task.cancelCause == nil {
		task.cancelCause = cancellationCause(task.ctx.Err())
	}
	_ = task.runtime.Cancel(ctx, task.taskID, task.cancelCause.Error())
}

func cancellationCause(cause error) error {
	if cause != nil {
		return cause
	}
	return context.Canceled
}

func stopTimer(t *time.Timer) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
}

func resetTimer(t *time.Timer, d time.Duration) {
	stopTimer(t)
	t.Reset(d)
}

func (a *Actor) nextPollInterval(claimed bool, inFlightCount int) time.Duration {
	if claimed || inFlightCount > 0 {
		return a.cfg.PollEvery
	}
	return a.cfg.IdlePollEvery
}

func (a *Actor) claimAvailable(ctx context.Context, runtimes []*runtimeactor.Actor, inFlight map[string]*runningTask) bool {
	claimCtx, release := a.beginClaim(ctx)
	defer release()

	claimed := false
	for _, rt := range runtimes {
		if claimCtx.Err() != nil {
			return claimed
		}
		status, err := rt.Status(claimCtx)
		if err != nil {
			continue
		}
		if a.claimOne(ctx, claimCtx, rt, status, inFlight) {
			claimed = true
		}
	}
	return claimed
}

func (a *Actor) beginClaim(ctx context.Context) (context.Context, func()) {
	claimCtx, cancel := context.WithCancel(ctx)
	a.claimMu.Lock()
	a.claimCancel = cancel
	a.claimMu.Unlock()

	return claimCtx, func() {
		cancel()
		a.claimMu.Lock()
		a.claimCancel = nil
		a.claimMu.Unlock()
	}
}

func (a *Actor) cancelCurrentClaim() {
	a.claimMu.Lock()
	cancel := a.claimCancel
	a.claimMu.Unlock()
	if cancel != nil {
		cancel()
	}
}
