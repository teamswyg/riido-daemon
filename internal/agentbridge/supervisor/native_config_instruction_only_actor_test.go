package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/policy"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func startNativeConfigInstructionOnlyActor(
	t *testing.T,
	provider bridge.Provider,
) (*reporterProbe, *process.FakeRunning) {
	t.Helper()

	source := controlplane.NewMemorySource()
	source.Enqueue(nativeConfigInstructionOnlyRequest(provider))
	reporter := newReporterProbe()
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	rt := startNamedRuntime(t, fake, "rt-"+string(provider), string(provider))
	actor, err := New(Config{
		DaemonID:            "daemon-1",
		RiidoDaemonVersion:  testRiidoDaemonVersion,
		Runtime:             rt,
		Source:              source,
		Reporter:            reporter,
		Workdir:             workdir.NewFSAdapter(t.TempDir()),
		PollEvery:           10 * time.Millisecond,
		HeartbeatEvery:      time.Hour,
		PolicyBundleVersion: policy.DefaultLocalPolicyBundleVersion,
		PolicyBundle:        policy.DefaultLocalPolicyBundle(),
		RuntimeTrustTier:    policy.TrustTierHost,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := actor.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = actor.Stop(ctx)
	})
	return reporter, running
}

func nativeConfigInstructionOnlyRequest(provider bridge.Provider) bridge.TaskRequest {
	return bridge.TaskRequest{
		ID:                       "t-" + string(provider) + "-native-config",
		Provider:                 provider,
		Prompt:                   "hello",
		AllowExperimentalRuntime: true,
		Metadata: map[string]string{
			MetadataWorkspaceID: "ws-1",
		},
	}
}
