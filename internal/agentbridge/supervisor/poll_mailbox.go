package supervisor

import "context"

func (a *Actor) handleMailboxMessage(
	ctx context.Context,
	msg envelope,
	inFlight map[string]*runningTask,
) bool {
	switch {
	case msg.taskActivation != nil:
		return a.handleTaskActivation(ctx, msg.taskActivation, inFlight)
	case msg.taskEvent != nil:
		a.handleTaskEvent(ctx, msg.taskEvent, inFlight)
		return false
	case msg.taskReport != nil:
		a.handleTaskReportEvent(ctx, msg.taskReport, inFlight)
		return false
	case msg.taskResult != nil:
		a.handleTaskResult(ctx, msg.taskResult, inFlight)
		return true
	case msg.cancel != nil:
		a.handleTaskCancel(ctx, msg.cancel, inFlight)
		return false
	default:
		return false
	}
}
