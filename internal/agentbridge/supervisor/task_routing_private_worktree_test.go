package supervisor

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func TestSupervisorStartsInIsolatedWorkspaceAfterPrivateCheckoutFailure(t *testing.T) {
	original := runAssignmentGitClone
	runAssignmentGitClone = func(_ context.Context, _ string, args []string) error {
		if err := os.WriteFile(filepath.Join(args[len(args)-1], "partial"), []byte("partial source"), 0o600); err != nil {
			return err
		}
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
	expectProcessStarted(t, running, "checkout failure prevented provider startup")
	completeFakeProcess(running)
	res := expectTaskResult(t, reporter, "isolated run did not finish")
	if res.Status != agentbridge.ResultCompleted || res.ExecutionContext == nil || res.ExecutionContext.WorkspaceKind != "isolated" {
		t.Fatalf("isolated result = %+v", res)
	}
	if _, err := os.Stat(filepath.Join(res.Workdir, "partial")); !os.IsNotExist(err) {
		t.Fatalf("partial checkout visible: %v", err)
	}
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
