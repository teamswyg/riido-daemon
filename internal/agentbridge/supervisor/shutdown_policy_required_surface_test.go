package supervisor

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestSupervisorBlocksTaskWhenRequiredSurfaceUnsupported(t *testing.T) {
	source := controlplane.NewMemorySource()
	source.Enqueue(unsupportedMCPSurfaceRequest())
	reporter := newReporterProbe()
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	startRoutingSupervisor(t, Config{
		DaemonID: "daemon-1",
		Runtime:  startRuntime(t, fake),
		Source:   source,
		Reporter: reporter,
	})

	expectStartedTask(t, reporter, "t-needs-mcp")
	res := expectTaskResult(t, reporter, "ineligible task was not reported")
	if res.Status != agentbridge.ResultBlocked {
		t.Fatalf("expected blocked result, got %+v", res)
	}
	if !strings.Contains(res.Error, "MISSING_REQUIRED_SURFACE:mcp") {
		t.Fatalf("missing scheduler reason: %+v", res)
	}
	if running.Command().Executable != "" {
		t.Fatalf("provider process should not have spawned: %+v", running.Command())
	}
}

func unsupportedMCPSurfaceRequest() bridge.TaskRequest {
	return bridge.TaskRequest{
		ID:               "t-needs-mcp",
		Provider:         "fake",
		Prompt:           "x",
		RequiredSurfaces: []string{"mcp"},
	}
}
