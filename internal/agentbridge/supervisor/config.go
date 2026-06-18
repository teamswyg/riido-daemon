package supervisor

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/policy"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

type Config struct {
	DaemonID string
	// RiidoDaemonVersion is the A-axis daemon binary version stamped on
	// CanonicalEvent common envelopes.
	RiidoDaemonVersion string
	// Runtime is the legacy single-runtime path used by tests and older
	// embedders. New daemon wiring should pass Runtimes, one RuntimeActor per
	// provider capability boundary.
	Runtime *runtimeactor.Actor
	// Runtimes is the provider-runtime pool the supervisor dispatches over.
	Runtimes []*runtimeactor.Actor
	Source   controlplane.TaskSourcePort
	Reporter controlplane.TaskReporterPort
	Workdir  workdir.Adapter

	PollEvery           time.Duration
	IdlePollEvery       time.Duration
	HeartbeatEvery      time.Duration
	MailboxSize         int
	PolicyBundleVersion string
	PolicyBundle        policy.PolicyBundle
	RuntimeTrustTier    policy.TrustTier
}
