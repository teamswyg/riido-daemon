package supervisor

import (
	"context"
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

func (a *Actor) blockPreparingRuntimeDrift(
	ctx context.Context,
	inFlight map[string]*runningTask,
	status runtimeactor.Status,
) {
	for _, task := range inFlight {
		if !preparingTaskDrifted(task, status) {
			continue
		}
		cancelRunningTask(task)
		reportCtx := controlplane.ContextWithTaskReport(ctx, task.report)
		_ = a.cfg.Reporter.CompleteTask(reportCtx, task.taskID, driftBlockedResult(task, status))
		delete(inFlight, task.taskID)
	}
}

func preparingTaskDrifted(task *runningTask, status runtimeactor.Status) bool {
	if task == nil || task.handle != nil || task.runtimeID != status.RuntimeID {
		return false
	}
	capView, ok := findCapability(status.Capabilities, task.provider)
	if !ok {
		return task.capabilityFingerprint != ""
	}
	return task.capabilityFingerprint != "" &&
		capView.CapabilityFingerprint != "" &&
		task.capabilityFingerprint != capView.CapabilityFingerprint
}

func driftBlockedResult(task *runningTask, status runtimeactor.Status) agentbridge.Result {
	return agentbridge.Result{
		Status: agentbridge.ResultBlocked,
		Error:  fmt.Sprintf("%v: task %s runtime %s drifted", runtimeactor.ErrRuntimePinViolated, task.taskID, status.RuntimeID),
	}
}
