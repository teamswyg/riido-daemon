package main

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/toolpolicy"
	"github.com/teamswyg/riido-daemon/internal/logging"
	"github.com/teamswyg/riido-daemon/internal/policy"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func daemonToolAutoApprover(settings daemonSettings) agentbridge.AutoApprover {
	return toolpolicy.PolicyAutoApprover(settings.PolicyBundleDoc, policy.TrustTierHost)
}

func daemonToolStartGate(settings daemonSettings) agentbridge.ToolStartGate {
	return toolpolicy.PolicyToolStartGate(settings.PolicyBundleDoc, policy.TrustTierHost)
}

func daemonToolApprovalGate(settings daemonSettings) agentbridge.ToolApprovalGate {
	return toolpolicy.PolicyToolApprovalGate(settings.PolicyBundleDoc, policy.TrustTierHost)
}

func daemonToolApprovalResolver(reporter controlplane.TaskReporterPort) agentbridge.ToolApprovalResolver {
	resolver, _ := reporter.(agentbridge.ToolApprovalResolver)
	return resolver
}

func daemonToolApprovalAuthorizer(reporter controlplane.TaskReporterPort) agentbridge.ToolApprovalAuthorizer {
	authorizer, _ := reporter.(agentbridge.ToolApprovalAuthorizer)
	return authorizer
}

func stopRuntimeActors(ctx lifecycle.Context, runtimes []*runtimeactor.Actor, log logging.Logger) {
	for _, rt := range runtimes {
		if err := rt.StopLifecycle(ctx); err != nil {
			log.Printf("runtimeactor stop error level=%s: %v", ctx.ShutdownLevel(), err)
		}
	}
}

func providerRuntimeID(daemonID, provider string) string {
	if provider == "" {
		return daemonID
	}
	return daemonID + ":" + provider
}
