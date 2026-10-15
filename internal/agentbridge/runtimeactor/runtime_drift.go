package runtimeactor

import (
	"context"
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func blockRuntimeDriftedTasks(ctx context.Context, inFlight map[string]*runningTask, newCap Capability) {
	if newCap.CapabilityFingerprint == "" {
		return
	}
	for _, task := range inFlight {
		if !taskPinnedToDriftedCapability(task, newCap) {
			continue
		}
		task.handle.session.TerminateWithContext(ctx, agentbridge.Result{
			Status: agentbridge.ResultBlocked,
			Error:  runtimePinViolationError(task, newCap),
		})
	}
}

func taskPinnedToDriftedCapability(task *runningTask, capView Capability) bool {
	return task != nil &&
		task.handle != nil &&
		task.provider == capView.Provider &&
		task.capabilityFingerprint != "" &&
		task.capabilityFingerprint != capView.CapabilityFingerprint
}

func runtimePinViolationError(task *runningTask, capView Capability) string {
	return fmt.Sprintf("%v: task %s provider %s fingerprint changed from %s to %s",
		ErrRuntimePinViolated,
		task.taskID,
		capView.Provider,
		task.capabilityFingerprint,
		capView.CapabilityFingerprint,
	)
}
