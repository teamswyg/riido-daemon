package supervisor

import (
	"path/filepath"
	"strings"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func TestSupervisorUsesLogicalTaskIDMetadataForWorkspace(t *testing.T) {
	gotCloneArgs := captureRunAssignmentGitClone(t)
	source := controlplane.NewMemorySource()
	source.Enqueue(logicalWorktreeRequest())
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
	expectStartedTask(t, reporter, "asn-1")
	completeFakeProcess(running)
	res := expectTaskResult(t, reporter, "result was not reported")
	wantSuffix := filepath.Join("ws-1", "tasks", "task-a", "runs", "asn-1", "workdir")
	if !strings.HasSuffix(res.Workdir, wantSuffix) {
		t.Fatalf("workdir = %q, want suffix %q", res.Workdir, wantSuffix)
	}
	if len(*gotCloneArgs) == 0 || (*gotCloneArgs)[len(*gotCloneArgs)-1] != res.Workdir {
		t.Fatalf("clone args did not target prepared workdir: args=%#v workdir=%q", *gotCloneArgs, res.Workdir)
	}
	events := readRunEvents(t, filepath.Join(filepath.Dir(res.Workdir), "ir", "events.jsonl"))
	assertRunEvent(t, events, ir.EventRunReportedDone, func(ev ir.CanonicalEvent) {
		if ev.TaskID != "task-a" || ev.RunID != "asn-1" {
			t.Fatalf("logical task/run ids not preserved: %+v", ev)
		}
	})
}

func logicalWorktreeRequest() bridge.TaskRequest {
	return bridge.TaskRequest{
		ID:                       "asn-1",
		Provider:                 "codex",
		Prompt:                   "hello",
		AllowExperimentalRuntime: true,
		Worktree: &assignmentcontract.AssignmentWorktree{
			RepositoryFullName: "teamswyg/riido-daemon",
			RepositoryURL:      "https://github.com/teamswyg/riido-daemon",
			BranchName:         "RIID-4964-agent-profile-upload",
		},
		Metadata: map[string]string{
			MetadataWorkspaceID:         "ws-1",
			MetadataRunID:               "asn-1",
			controlplane.MetadataTaskID: "task-a",
		},
	}
}
