package supervisor

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func (a *Actor) run(ctx context.Context, runtimes []*runtimeactor.Actor) {
	defer close(a.stoppedCh)
	heartbeatCtx, stopHeartbeats := context.WithCancel(ctx)
	heartbeatDone := make(chan struct{})
	go func() { defer close(heartbeatDone); a.heartbeatLoop(heartbeatCtx, runtimes) }()
	defer func() { stopHeartbeats(); <-heartbeatDone }()
	poll := time.NewTimer(a.cfg.PollEvery)
	defer stopTimer(poll)
	heartbeat := time.NewTicker(a.cfg.HeartbeatEvery)
	defer heartbeat.Stop()

	var assignmentNotices <-chan struct{}
	if source, ok := a.cfg.Source.(interface {
		AssignmentNotifications(context.Context) <-chan struct{}
	}); ok {
		assignmentNotices = source.AssignmentNotifications(heartbeatCtx)
	}
	inFlight := map[string]*runningTask{}
	detachedReports := []detachedReport{}
	var quiesceUntil time.Time

	for {
		a.publishHeartbeatExecutions(inFlight)
		select {
		case <-ctx.Done():
			level := lifecycle.FromContext(ctx).ShutdownLevel()
			a.stopRun(runtimes, inFlight, &detachedReports, level, ctx.Err())
			return
		case level := <-a.stopReqCh:
			a.stopRun(runtimes, inFlight, &detachedReports, a.drainStopLevel(level), nil)
			return
		case request := <-a.idleStops:
			ready := request.ctx.Err() == nil && idleForUpdate(inFlight, detachedReports, len(a.mailbox))
			if ready {
				quiesceUntil = time.Now().Add(10 * time.Second)
			}
			request.reply <- ready
		case <-assignmentNotices:
			resetTimer(poll, time.Millisecond)
		case <-poll.C:
			a.retryEventReports(ctx, inFlight)
			a.retryTerminalReports(ctx, inFlight)
			a.retryDetachedReports(ctx, &detachedReports)
			claimed := false
			if !time.Now().Before(quiesceUntil) {
				claimed = a.claimAvailable(ctx, runtimes, inFlight)
			}
			resetTimer(poll, a.nextPollInterval(claimed, len(inFlight)))
		case <-heartbeat.C:
			a.refreshRuntimeRegistrations(ctx, runtimes, inFlight)
		case msg := <-a.mailbox:
			if a.handleMailboxDelivery(ctx, msg, inFlight, &detachedReports) {
				resetTimer(poll, a.cfg.PollEvery)
			}
		}
	}
}

func (a *Actor) stopRun(
	runtimes []*runtimeactor.Actor,
	inFlight map[string]*runningTask,
	detachedReports *[]detachedReport,
	level lifecycle.ShutdownLevel,
	err error,
) {
	shutdownCtx, cancel := supervisorShutdownContext(level)
	a.shutdown(shutdownCtx, runtimes, inFlight, detachedReports)
	cancel()
	a.stopErrCh <- err
}
