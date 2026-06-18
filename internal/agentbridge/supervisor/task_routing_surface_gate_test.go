package supervisor

import (
	"strings"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestSupervisorBlocksWorktreeWhenRuntimeSurfaceUnsupported(t *testing.T) {
	source := controlplane.NewMemorySource()
	source.Enqueue(unsupportedWorktreeRequest())
	reporter := newReporterProbe()
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	rt := startRuntime(t, fake)

	startRoutingSupervisor(t, Config{DaemonID: "daemon-1", Runtime: rt, Source: source, Reporter: reporter})
	expectStartedTask(t, reporter, "asn-worktree-unsupported")
	res := expectTaskResult(t, reporter, "blocked result was not reported")
	if res.Status != agentbridge.ResultBlocked {
		t.Fatalf("result status = %s, want %s; result=%+v", res.Status, agentbridge.ResultBlocked, res)
	}
	if !strings.Contains(res.Error, "MISSING_REQUIRED_SURFACE:worktree") {
		t.Fatalf("result error should explain missing worktree surface: %q", res.Error)
	}
	assertProcessDoesNotStart(t, running, "provider process should not start when worktree surface is unsupported")
}

func unsupportedWorktreeRequest() bridge.TaskRequest {
	return bridge.TaskRequest{
		ID:       "asn-worktree-unsupported",
		Provider: "fake",
		Prompt:   "fix repo",
		Worktree: &assignmentcontract.AssignmentWorktree{
			RepositoryFullName: "teamswyg/riido-daemon",
			RepositoryURL:      "https://github.com/teamswyg/riido-daemon",
			BranchName:         "main",
		},
	}
}
