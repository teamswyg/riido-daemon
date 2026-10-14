package supervisor

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func TestSupervisorOnlyLongPollsLastRuntimeInClaimWave(t *testing.T) {
	source := newClaimLongPollProbe()
	rtClaude := startNamedRuntime(t, process.NewFake(), "rt-claude", "claude")
	rtCodex := startNamedRuntime(t, process.NewFake(), "rt-codex", "codex")
	startRoutingSupervisor(t, Config{
		DaemonID:      "daemon-1",
		Runtimes:      []*runtimeactor.Actor{rtClaude, rtCodex},
		Source:        source,
		Reporter:      newReporterProbe(),
		Workdir:       workdir.NewFSAdapter(t.TempDir()),
		IdlePollEvery: time.Hour,
	})

	first := source.expectClaim(t)
	second := source.expectClaim(t)
	if first.runtimeID != "rt-claude" || first.longPoll {
		t.Fatalf("first claim = %+v, want rt-claude without long poll", first)
	}
	if second.runtimeID != "rt-codex" || !second.longPoll {
		t.Fatalf("second claim = %+v, want rt-codex with long poll", second)
	}
}
