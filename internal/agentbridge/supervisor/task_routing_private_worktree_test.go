package supervisor

import (
	"strings"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func TestSupervisorBlocksPrivateAssignmentWorktreeBeforeProviderStart(t *testing.T) {
	source := controlplane.NewMemorySource()
	source.Enqueue(privateWorktreeRequest())
	reporter := newReporterProbe()
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	rt := startNamedRuntime(t, fake, "rt-codex", "codex")

	startRoutingSupervisor(t, Config{
		DaemonID: "daemon-1",
		Runtime:  rt,
		Source:   source,
		Reporter: reporter,
		Workdir:  workdir.NewFSAdapter(t.TempDir()),
	})
	expectStartedTask(t, reporter, "asn-private")
	res := expectTaskResult(t, reporter, "blocked result was not reported")
	if res.Status != agentbridge.ResultBlocked {
		t.Fatalf("result status = %s, want %s; result=%+v", res.Status, agentbridge.ResultBlocked, res)
	}
	if !strings.Contains(res.Error, "private assignment worktree requires git credentials") {
		t.Fatalf("result error does not explain private worktree credentials: %q", res.Error)
	}
	assertProcessDoesNotStart(t, running, "provider process should not start for blocked private worktree")
}

func privateWorktreeRequest() bridge.TaskRequest {
	return bridge.TaskRequest{
		ID:                       "asn-private",
		Provider:                 "codex",
		Prompt:                   "fix private repo",
		AllowExperimentalRuntime: true,
		Worktree: &assignmentcontract.AssignmentWorktree{
			RepositoryFullName: "teamswyg/private-repo",
			RepositoryURL:      "https://github.com/teamswyg/private-repo",
			BranchName:         "main",
			IsPrivate:          true,
		},
		Metadata: map[string]string{
			MetadataWorkspaceID:         "ws-1",
			MetadataRunID:               "asn-private",
			controlplane.MetadataTaskID: "task-private",
		},
	}
}
