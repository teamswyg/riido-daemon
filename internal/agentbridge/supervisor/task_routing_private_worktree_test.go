package supervisor

import (
	"context"
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
	original := runAssignmentGitClone
	runAssignmentGitClone = func(context.Context, string, []string) error {
		return classifyAssignmentGitCloneFailure([]byte("fatal: Authentication failed for https://fixture-secret@github.com"))
	}
	t.Cleanup(func() { runAssignmentGitClone = original })
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
	if !strings.Contains(res.Error, "Git authentication was rejected") || !strings.Contains(res.Error, "gh auth setup-git") || strings.Contains(res.Error, "fixture-secret") {
		t.Fatalf("result error must explain the private worktree failure and recovery: %q", res.Error)
	}
	assertProcessDoesNotStart(t, running, "provider process should not start for blocked private worktree")
}

func TestSupervisorStartsPrivateAssignmentAfterGitClone(t *testing.T) {
	clones := captureRunAssignmentGitClone(t)
	source := controlplane.NewMemorySource()
	source.Enqueue(privateWorktreeRequest())
	reporter := newReporterProbe()
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	rt := startNamedRuntime(t, fake, "rt-codex", "codex")
	startRoutingSupervisor(t, Config{
		DaemonID: "daemon-1", Runtime: rt, Source: source, Reporter: reporter,
		Workdir: workdir.NewFSAdapter(t.TempDir()),
	})
	expectStartedTask(t, reporter, "asn-private")
	completeFakeProcess(running)
	res := expectTaskResult(t, reporter, "private task did not finish")
	if res.Status != agentbridge.ResultCompleted || len(*clones) == 0 {
		t.Fatalf("private task result = %+v, clone arguments = %v", res, *clones)
	}
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
