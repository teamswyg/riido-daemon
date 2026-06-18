package supervisor

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/process"
)

type idlePollingRun struct {
	source *idlePollSource
}

func startIdlePollingSupervisor(t *testing.T) idlePollingRun {
	t.Helper()
	source := newIdlePollSource()
	rt := startRuntime(t, process.NewFake())
	startRoutingSupervisor(t, Config{
		DaemonID:       "daemon-1",
		Runtime:        rt,
		Source:         source,
		Reporter:       newReporterProbe(),
		PollEvery:      10 * time.Millisecond,
		IdlePollEvery:  120 * time.Millisecond,
		HeartbeatEvery: time.Hour,
	})
	return idlePollingRun{source: source}
}
