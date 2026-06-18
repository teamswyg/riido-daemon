package supervisor

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func TestSupervisorWorkdirRequiresWorkspaceID(t *testing.T) {
	source := controlplane.NewMemorySource()
	source.Enqueue(bridge.TaskRequest{ID: "t-no-workspace", Provider: "fake", Prompt: "x"})
	reporter := newReporterProbe()
	startRoutingSupervisor(t, Config{
		DaemonID: "daemon-1",
		Runtime:  startRuntime(t, process.NewFake()),
		Source:   source,
		Reporter: reporter,
		Workdir:  workdir.NewFSAdapter(t.TempDir()),
	})

	expectStartedTask(t, reporter, "t-no-workspace")
	res := expectTaskResult(t, reporter, "workdir failure was not reported")
	if res.Status != agentbridge.ResultFailed || res.Error == "" {
		t.Fatalf("expected workdir failure result, got %+v", res)
	}
}
