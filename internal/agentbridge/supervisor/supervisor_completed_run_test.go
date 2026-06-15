package supervisor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func assertSupervisorCompletedRun(t *testing.T, res agentbridge.Result, running *process.FakeRunning) {
	t.Helper()
	if res.Status != agentbridge.ResultCompleted || res.Output != "done" {
		t.Fatalf("result: %+v", res)
	}
	if res.Workdir == "" {
		t.Fatalf("expected isolated workdir in result: %+v", res)
	}
	if running.Command().Dir != res.Workdir {
		t.Fatalf("spawn dir %q != result workdir %q", running.Command().Dir, res.Workdir)
	}
	if !hasEnvPrefix(running.Command().Env, "TEST_NATIVE_CONFIG_VERSION=") {
		t.Fatalf("native config version was not passed to adapter metadata: %+v", running.Command())
	}
	assertNativeConfigInjected(t, res.Workdir)
	assertArchiveManifest(t, res.Workdir)
	assertCompletedRunEvents(t, res.Workdir)
}

func assertNativeConfigInjected(t *testing.T, runWorkdir string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(runWorkdir, "AGENTS.md")); err != nil {
		t.Fatalf("runtime config not injected: %v", err)
	}
	nativeConfig, err := os.ReadFile(filepath.Join(runWorkdir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read runtime config: %v", err)
	}
	if !strings.Contains(string(nativeConfig), "<riido_log>") {
		t.Fatalf("runtime config missing telemetry hard rule:\n%s", nativeConfig)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(runWorkdir), "native-config", "AGENTS.md")); err != nil {
		t.Fatalf("native config copy not injected: %v", err)
	}
	manifest := readNativeConfigManifest(t, filepath.Join(runWorkdir, workdir.NativeConfigManifestPath))
	if manifest.ProviderKind != "fake" ||
		manifest.ProtocolKind != "fake-unknown" ||
		manifest.PrimaryInstructionFile != "AGENTS.md" ||
		manifest.TelemetryContractPlacement != agentbridge.TelemetryPlacementPrompt ||
		manifest.HookMode != workdir.NativeConfigHookModeInstructionOnly {
		t.Fatalf("native config manifest = %+v", manifest)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(runWorkdir), "native-config", filepath.FromSlash(workdir.NativeConfigManifestPath))); err != nil {
		t.Fatalf("native config manifest copy not injected: %v", err)
	}
}

func assertArchiveManifest(t *testing.T, runWorkdir string) {
	t.Helper()
	archive, err := os.ReadFile(filepath.Join(filepath.Dir(runWorkdir), "archive.json"))
	if err != nil {
		t.Fatalf("archive manifest not written: %v", err)
	}
	for _, want := range []string{`"schema_version": "riido-workdir-archive.v1"`, `"retention_mode": "keep-in-place"`, `"result_status": "completed"`} {
		if !strings.Contains(string(archive), want) {
			t.Fatalf("archive manifest missing %q:\n%s", want, archive)
		}
	}
}

func assertCompletedRunEvents(t *testing.T, runWorkdir string) {
	t.Helper()
	events := readRunEvents(t, filepath.Join(filepath.Dir(runWorkdir), "ir", "events.jsonl"))
	assertRunEvent(t, events, ir.EventWorkdirCreated, func(ev ir.CanonicalEvent) {
		if ev.NativeConfigVersion != "" {
			t.Fatalf("WorkdirCreated must remain pre-execute without NCV: %+v", ev)
		}
		if ev.RiidoDaemonVersion != "riido-agentd v1.2.3" {
			t.Fatalf("daemon version not stamped: %+v", ev)
		}
	})
	assertRunEvent(t, events, ir.EventNativeConfigInjected, func(ev ir.CanonicalEvent) {
		if ev.NativeConfigVersion == "" {
			t.Fatalf("NativeConfigInjected missing NCV: %+v", ev)
		}
	})
	assertCompletedProviderEvent(t, events)
	assertRunEvent(t, events, ir.EventWorkdirArchived, func(ev ir.CanonicalEvent) {
		if ev.NativeConfigVersion == "" {
			t.Fatalf("WorkdirArchived missing NCV: %+v", ev)
		}
	})
}

func assertCompletedProviderEvent(t *testing.T, events []ir.CanonicalEvent) {
	t.Helper()
	assertRunEvent(t, events, ir.EventTextDelta, func(ev ir.CanonicalEvent) {
		if ev.NativeConfigVersion == "" {
			t.Fatalf("TextDelta missing NCV: %+v", ev)
		}
		if ev.ActorKind != ir.ActorAgent || ev.ActorID != "t-1" {
			t.Fatalf("provider event attribution mismatch: %+v", ev)
		}
		if ev.Payload["text"] != "done" {
			t.Fatalf("TextDelta payload mismatch: %+v", ev.Payload)
		}
	})
	assertRunEvent(t, events, ir.EventRunReportedDone, func(ev ir.CanonicalEvent) {
		if ev.NativeConfigVersion == "" {
			t.Fatalf("RunReportedDone missing NCV: %+v", ev)
		}
		if ev.FSMVersion != task.FSMSchemaVersion {
			t.Fatalf("RunReportedDone FSMVersion = %d, want %d", ev.FSMVersion, task.FSMSchemaVersion)
		}
		if ev.ActorKind != ir.ActorDaemon {
			t.Fatalf("RunReportedDone must be daemon-attributed: %+v", ev)
		}
	})
}
