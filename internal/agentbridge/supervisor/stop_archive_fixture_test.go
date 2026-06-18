package supervisor

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

const stopArchiveTaskID = "t-stop"

type stopArchiveFixture struct {
	actor    *Actor
	reporter *reporterProbe
	running  *process.FakeRunning
}

func startStopArchiveFixture(t *testing.T) stopArchiveFixture {
	t.Helper()
	source := controlplane.NewMemorySource()
	source.Enqueue(stopArchiveRequest())

	reporter := newReporterProbe()
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running

	return stopArchiveFixture{
		actor:    startStopArchiveSupervisor(t, source, reporter, fake),
		reporter: reporter,
		running:  running,
	}
}

func stopArchiveRequest() bridge.TaskRequest {
	return bridge.TaskRequest{
		ID:       stopArchiveTaskID,
		Provider: "fake",
		Prompt:   "x",
		Metadata: map[string]string{
			MetadataWorkspaceID: "ws-stop",
			MetadataRunID:       "run-stop",
		},
	}
}

func startStopArchiveSupervisor(
	t *testing.T,
	source *controlplane.MemorySource,
	reporter *reporterProbe,
	fake *process.Fake,
) *Actor {
	t.Helper()
	return startRoutingSupervisor(t, Config{
		DaemonID:           "daemon-stop",
		RiidoDaemonVersion: "riido-agentd v1.2.3",
		Runtime:            startRuntime(t, fake),
		Source:             source,
		Reporter:           reporter,
		Workdir:            workdir.NewFSAdapter(t.TempDir()),
		PollEvery:          10 * time.Millisecond,
		HeartbeatEvery:     time.Hour,
	})
}
