package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/policy"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

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
