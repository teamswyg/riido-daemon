package supervisor

import (
	"context"
	"time"

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
				}
				_ = a.cfg.Reporter.CompleteTask(reportCtx, msg.taskResult.taskID, res)
				delete(inFlight, msg.taskResult.taskID)
				resetTimer(poll, a.cfg.PollEvery)
			case msg.cancel != nil:
				if inFlight[msg.cancel.taskID] != nil {
					reason := "cancelled"
					if msg.cancel.cause != nil {
						reason = msg.cancel.cause.Error()
					}
					_ = inFlight[msg.cancel.taskID].runtime.Cancel(ctx, msg.cancel.taskID, reason)
				}
			}
		}
	}
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
