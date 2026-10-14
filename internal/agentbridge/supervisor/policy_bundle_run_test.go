package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func runPolicyBundleScenario(t *testing.T, cfg policyBundleScenarioConfig) policyBundleScenario {
	t.Helper()
	source := controlplane.NewMemorySource()
	source.Enqueue(bridge.TaskRequest{
		ID:                       cfg.taskID,
		Provider:                 bridge.Provider(cfg.provider),
		Prompt:                   "hello",
		AllowExperimentalRuntime: cfg.allowExperimentalRuntime,
		Metadata: map[string]string{
			MetadataWorkspaceID: "ws-1",
		},
	})

	reporter := newReporterProbe()
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	bundle := noSurfacePolicyBundle(cfg.bundleVersion)
	rt := startNamedRuntime(t, fake, "rt-"+cfg.provider, cfg.provider)
	actor := startPolicyBundleActor(t, source, reporter, rt, bundle)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = actor.Stop(ctx)
	})
	waitPolicyBundleTaskClaimed(t, reporter)
	return policyBundleScenario{
		running: running,
		result:  completePolicyBundleTask(t, reporter, running),
	}
}
