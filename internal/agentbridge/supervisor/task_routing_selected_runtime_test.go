package supervisor

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func TestSupervisorDispatchesTaskToSelectedRuntimeActor(t *testing.T) {
	source := newRuntimeRoutingSource(selectedRuntimeClaims())
	reporter := newReporterProbe()
	claudeFake := process.NewFake()
	codexFake := process.NewFake()
	codexRunning := process.NewFakeRunning()
	codexFake.NextRunning = codexRunning
	rtClaude := startNamedRuntime(t, claudeFake, "rt-claude", "claude")
	rtCodex := startNamedRuntime(t, codexFake, "rt-codex", "codex")

	startRoutingSupervisor(t, Config{
		DaemonID: "daemon-1",
		Runtimes: []*runtimeactor.Actor{rtClaude, rtCodex},
		Source:   source,
		Reporter: reporter,
		Workdir:  workdir.NewFSAdapter(t.TempDir()),
	})
	assertRuntimeRegistrations(t, source)
	expectStartedTask(t, reporter, "t-codex")
	select {
	case cmd := <-codexRunning.StartedRecv():
		if cmd.Executable != "codex" {
			t.Fatalf("codex runtime command mismatch: %+v", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("codex runtime did not spawn process")
	}
	completeFakeProcess(codexRunning)
	res := expectTaskResult(t, reporter, "result was not reported")
	if res.Status != agentbridge.ResultCompleted {
		t.Fatalf("result: %+v", res)
	}
}

func selectedRuntimeClaims() map[string][]bridge.TaskRequest {
	return map[string][]bridge.TaskRequest{"rt-codex": {{
		ID:                       "t-codex",
		Provider:                 "codex",
		Prompt:                   "hello",
		AllowExperimentalRuntime: true,
		Metadata:                 map[string]string{MetadataWorkspaceID: "ws-1"},
	}}}
}
