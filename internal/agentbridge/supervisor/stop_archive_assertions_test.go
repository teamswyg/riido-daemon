package supervisor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
)

func assertStopArchiveManifest(t *testing.T, runWorkdir string) {
	t.Helper()
	runRoot := filepath.Dir(runWorkdir)
	archive, err := os.ReadFile(filepath.Join(runRoot, "archive.json"))
	if err != nil {
		t.Fatalf("archive manifest not written on stop: %v", err)
	}
	if !strings.Contains(string(archive), `"result_status": "cancelled"`) {
		t.Fatalf("archive manifest should record cancelled status:\n%s", archive)
	}
}

func assertStopArchiveRunEvents(t *testing.T, runWorkdir string) {
	t.Helper()
	runRoot := filepath.Dir(runWorkdir)
	events := readRunEvents(t, filepath.Join(runRoot, "ir", "events.jsonl"))
	assertRunEvent(t, events, ir.EventTaskCancelled, func(event ir.CanonicalEvent) {
		assertStopArchiveCancelledEvent(t, event)
	})
	assertRunEvent(t, events, ir.EventWorkdirArchived, nil)
}

func assertStopArchiveCancelledEvent(t *testing.T, event ir.CanonicalEvent) {
	t.Helper()
	if event.ActorKind != ir.ActorDaemon {
		t.Fatalf("TaskCancelled must be daemon-attributed: %+v", event)
	}
	if event.FSMVersion != task.FSMSchemaVersion {
		t.Fatalf("TaskCancelled FSMVersion = %d, want %d", event.FSMVersion, task.FSMSchemaVersion)
	}
}
