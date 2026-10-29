package main

import (
	"os"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/process/processexec"
)

func newDaemonRuntimeActors(settings daemonSettings, adapters []agentbridge.Adapter, resolver agentbridge.ToolApprovalResolver) ([]*runtimeactor.Actor, error) {
	out := make([]*runtimeactor.Actor, 0, len(adapters))
	for _, adapter := range adapters {
		rt, err := newDaemonRuntimeActorForAdapter(settings, adapter, resolver)
		if err != nil {
			return nil, err
		}
		out = append(out, rt)
	}
	if len(out) == 0 {
		return nil, daemonErrorf(ErrDaemonRuntime, "runtime.new", "runtimeactor.New: at least one adapter is required")
	}
	return out, nil
}

func newDaemonRuntimeActorForAdapter(settings daemonSettings, adapter agentbridge.Adapter, resolver agentbridge.ToolApprovalResolver) (*runtimeactor.Actor, error) {
	name := strings.TrimSpace(adapter.Name())
	if name == "" {
		return nil, daemonErrorf(ErrDaemonRuntime, "runtime.new", "runtimeactor.New: adapter name is required")
	}
	rt, err := newDaemonRuntimeActor(settings, providerRuntimeID(settings.DaemonID, name), adapter, settings.RuntimeAgents, resolver)
	if err != nil {
		return nil, daemonWrapf(ErrDaemonRuntime, "runtime.new", err, "runtimeactor.New(%s)", name)
	}
	return rt, nil
}

func newDaemonRuntimeActor(settings daemonSettings, runtimeID string, adapter agentbridge.Adapter, agents []runtimeactor.AgentStatus, resolver agentbridge.ToolApprovalResolver) (*runtimeactor.Actor, error) {
	cliExecutable := ""
	if settings.contextBroker != nil {
		var err error
		cliExecutable, err = os.Executable()
		if err != nil {
			return nil, daemonWrapf(ErrDaemonRuntime, "runtime.context-cli-executable", err, "resolve daemon executable")
		}
	}
	return runtimeactor.New(runtimeactor.Config{
		RuntimeID:   runtimeID,
		HardTimeout: settings.RunHardTimeout, SemanticIdle: settings.RunSemanticIdle,
		Owner:                settings.RuntimeOwner,
		DeviceName:           settings.DeviceName,
		Agents:               agents,
		Models:               daemonRuntimeModels(adapter.Name()),
		RefreshModels:        refreshDaemonRuntimeModels,
		Adapters:             []agentbridge.Adapter{adapter},
		Process:              processexec.New(),
		MaxConcurrent:        settings.RuntimeMaxConcurrent,
		AutoApprove:          daemonToolAutoApprover(settings),
		ToolStartGate:        daemonToolStartGate(settings),
		ToolApprovalGate:     daemonToolApprovalGate(settings),
		ToolApprovalResolver: resolver,
		PolicyBundleVersion:  settings.PolicyBundle,
		LaunchEnv:            daemonContextLaunchEnv(settings.contextBroker),
		ContextCLIExecutable: cliExecutable,
	})
}
