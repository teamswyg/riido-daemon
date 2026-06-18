package supervisor

import (
	"context"
	"time"

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

	for {
		select {
		case <-ctx.Done():
			level := lifecycle.FromContext(ctx).ShutdownLevel()
			a.stopRun(runtimes, inFlight, level, ctx.Err())
			return
		case level := <-a.stopReqCh:
			a.stopRun(runtimes, inFlight, a.drainStopLevel(level), nil)
			return
		case <-poll.C:
			claimed := a.claimAvailable(ctx, runtimes, inFlight)
			resetTimer(poll, a.nextPollInterval(claimed, len(inFlight)))
		case <-heartbeat.C:
			a.reportRuntimeHeartbeats(ctx, runtimes)
		case msg := <-a.mailbox:
			if a.handleMailboxMessage(ctx, msg, inFlight) {
				resetTimer(poll, a.cfg.PollEvery)
			}
		}
	}
}

func (a *Actor) stopRun(
	runtimes []*runtimeactor.Actor,
	inFlight map[string]*runningTask,
	level lifecycle.ShutdownLevel,
	err error,
) {
	shutdownCtx, cancel := supervisorShutdownContext(level)
	a.shutdown(shutdownCtx, runtimes, inFlight)
	cancel()
	a.stopErrCh <- err
}
