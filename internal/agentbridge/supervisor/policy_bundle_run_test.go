package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/policy"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
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

func startPolicyBundleActor(
	t *testing.T,
	source *controlplane.MemorySource,
	reporter *reporterProbe,
	rt *runtimeactor.Actor,
	bundle policy.PolicyBundle,
) *Actor {
	t.Helper()
	actor, err := New(Config{
		DaemonID:            "daemon-1",
		Runtime:             rt,
		Source:              source,
		Reporter:            reporter,
		Workdir:             workdir.NewFSAdapter(t.TempDir()),
		PollEvery:           10 * time.Millisecond,
		HeartbeatEvery:      time.Hour,
		PolicyBundleVersion: bundle.Version,
		PolicyBundle:        bundle,
		RuntimeTrustTier:    policy.TrustTierHost,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := actor.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	return actor
}
