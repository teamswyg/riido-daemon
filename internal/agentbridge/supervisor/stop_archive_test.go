package supervisor

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func TestSupervisorStopArchivesInFlightWorkspace(t *testing.T) {
	source := controlplane.NewMemorySource()
	source.Enqueue(bridge.TaskRequest{
		ID:       "t-stop",
		Provider: "fake",
		Prompt:   "x",
		Metadata: map[string]string{
			MetadataWorkspaceID: "ws-stop",
			MetadataRunID:       "run-stop",
		},
	})

	reporter := newReporterProbe()
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	rt := startRuntime(t, fake)

	actor, err := New(Config{
		DaemonID:           "daemon-stop",
		RiidoDaemonVersion: "riido-agentd v1.2.3",
		Runtime:            rt,
		Source:             source,
		Reporter:           reporter,
		Workdir:            workdir.NewFSAdapter(t.TempDir()),
		PollEvery:          10 * time.Millisecond,
		HeartbeatEvery:     time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := actor.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	select {
	case <-reporter.started:
	case <-time.After(2 * time.Second):
		t.Fatal("task was not claimed")
	}

	var cmd process.Command
	select {
	case cmd = <-running.StartedRecv():
		if cmd.Dir == "" {
			t.Fatalf("provider command missing workdir: %+v", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("provider process was not spawned")
	}

	running.EmitStdout([]byte("event"))
	select {
	case ev := <-reporter.events:
		if ev.Kind != agentbridge.EventLifecycle || ev.Phase != agentbridge.StateRunning {
			t.Fatalf("running event: %+v", ev)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("running event was not reported")
	}

	select {
	case ev := <-reporter.events:
		if ev.Kind != agentbridge.EventTextDelta || ev.Text != "event" {
			t.Fatalf("event: %+v", ev)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("nonterminal event was not reported")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := actor.Stop(ctx); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	select {
	case <-running.KillRecv():
	case <-time.After(2 * time.Second):
		t.Fatal("provider process was not killed on supervisor stop")
	}

	var res agentbridge.Result
	select {
	case res = <-reporter.results:
		if res.Status != agentbridge.ResultCancelled || !strings.Contains(res.Error, "supervisor: stopped") {
			t.Fatalf("shutdown result: %+v", res)
		}
		if res.Workdir != cmd.Dir {
			t.Fatalf("shutdown result workdir = %q, want %q", res.Workdir, cmd.Dir)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("shutdown result was not reported")
	}

	runRoot := filepath.Dir(res.Workdir)
	archive, err := os.ReadFile(filepath.Join(runRoot, "archive.json"))
	if err != nil {
		t.Fatalf("archive manifest not written on stop: %v", err)
	}
	if !strings.Contains(string(archive), `"result_status": "cancelled"`) {
		t.Fatalf("archive manifest should record cancelled status:\n%s", archive)
	}

	events := readRunEvents(t, filepath.Join(runRoot, "ir", "events.jsonl"))
	assertRunEvent(t, events, ir.EventTaskCancelled, func(ev ir.CanonicalEvent) {
		if ev.ActorKind != ir.ActorDaemon {
			t.Fatalf("TaskCancelled must be daemon-attributed: %+v", ev)
		}
		if ev.FSMVersion != task.FSMSchemaVersion {
			t.Fatalf("TaskCancelled FSMVersion = %d, want %d", ev.FSMVersion, task.FSMSchemaVersion)
		}
	})
	assertRunEvent(t, events, ir.EventWorkdirArchived, nil)
}
