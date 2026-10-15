package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (a *Actor) handleTaskActivation(
	ctx context.Context,
	msg *taskActivationMsg,
	inFlight map[string]*runningTask,
) bool {
	task := inFlight[msg.taskID]
	if task == nil || task.terminalResult != nil || (msg.claimContext != nil && task.ctx != msg.claimContext) {
		if msg.handle != nil {
			// A cancelled old run already owns its own execution context. Drain
			// only that handle; cancelling by task ID could touch a newer claim.
			go func() {
				for range msg.handle.Events() {
				}
				<-msg.handle.Result()
			}()
		}
		return false
	}
	applyActivationRecovery(task, msg)
	if msg.runtime != nil {
		task.runtime = msg.runtime
		// Assignment lease ownership stays on the originally claimed runtime.
		if msg.prepared != nil && msg.prepared.request != nil {
			task.provider = string(msg.prepared.request.Provider)
		}
		if capView, ok := findCapability(msg.status.Capabilities, task.provider); ok {
			task.capabilityFingerprint = capView.CapabilityFingerprint
		}
	}
	if msg.err != nil {
		a.finishActivationError(ctx, task, msg, inFlight)
		return true
	}
	if msg.handle == nil {
		a.finishActivationMissingHandle(ctx, task, inFlight)
		return true
	}
	task.handle = msg.handle
	applyPreparedWorkspace(task, msg.prepared)
	if task.cancelCause != nil || task.ctx.Err() != nil {
		a.cancelActivatedTask(ctx, task)
		forwardActivatedSession(a, task)
		return false
	}
	if task.recovery != nil {
		a.reportTaskEvent(ctx, task, agentbridge.Event{Kind: agentbridge.EventProgress, Text: "Provider recovery: dispatched", RecoveryContext: task.recovery})
	}
	executionContext := executionContextForTask(task)
	a.reportTaskEvent(ctx, task, agentbridge.Event{
		ExecutionContext: &executionContext,

		Kind:  agentbridge.EventLifecycle,
		Phase: agentbridge.StateRunning,
	})
	forwardActivatedSession(a, task)
	return false
}

func applyPreparedWorkspace(task *runningTask, prepared *preparedWorkspace) {
	if prepared == nil {
		return
	}
	task.workspace = prepared.workspace
	task.execution = prepared.execution
	task.request = prepared.request
	task.canResume = prepared.canResume
	task.executionVerified = false
	task.events = prepared.events
}

func forwardActivatedSession(a *Actor, task *runningTask) {
	go a.forwardSession(task.taskID, task.handle.Events(), task.handle.Result())
}

// Correction activation reuses the original recovery ownership and provider
// boundary unless a new pre-dispatch recovery explicitly replaces them.
func applyActivationRecovery(task *runningTask, msg *taskActivationMsg) {
	if msg.recovery != nil {
		task.recovery = msg.recovery
		task.recoveryScope = msg.recoveryScope
	}
}
