package supervisor

import "context"

func (a *Actor) handleTaskCancel(
	ctx context.Context,
	msg *cancelMsg,
	inFlight map[string]*runningTask,
) {
	task := inFlight[msg.taskID]
	if task == nil {
		return
	}
	task.cancelCause = cancellationCause(msg.cause)
	cancelRunningTask(task)
	if task.handle != nil {
		_ = task.runtime.Cancel(ctx, task.taskID, task.cancelCause.Error())
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
