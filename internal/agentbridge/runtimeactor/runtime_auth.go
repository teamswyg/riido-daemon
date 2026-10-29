package runtimeactor

import (
	"context"
	"fmt"

	"github.com/teamswyg/riido-contracts/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func blockUnauthenticatedTasks(ctx context.Context, inFlight map[string]*runningTask, observed Capability) {
	if observed.Available || observed.HealthStatus != hostintegration.ProviderHealthUnavailable ||
		observed.DiagnosticCode != hostintegration.ProviderDiagnosticLoginRequired {
		return
	}
	for _, task := range inFlight {
		// A changed identity is fenced separately and retains the drift reason.
		if task == nil || task.handle == nil || task.provider != observed.Provider ||
			task.capabilityFingerprint == "" || task.capabilityFingerprint != observed.CapabilityFingerprint {
			continue
		}
		task.handle.session.TerminateWithContext(ctx, agentbridge.Result{
			Status: agentbridge.ResultBlocked,
			Error:  fmt.Sprintf("runtimeactor: provider %s authentication required", observed.Provider),
		})
	}
}
