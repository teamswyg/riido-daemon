package supervisor

import (
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
)

func assertCompletedRunEvents(t *testing.T, runWorkdir string) {
	t.Helper()

	events := readRunEvents(t, completedRunEventLogPath(runWorkdir))
	assertCompletedWorkdirCreatedEvent(t, events)
	assertCompletedNativeConfigInjectedEvent(t, events)
	assertCompletedProviderEvent(t, events)
	assertCompletedWorkdirArchivedEvent(t, events)
}

func completedRunEventLogPath(runWorkdir string) string {
	return filepath.Join(filepath.Dir(runWorkdir), "ir", "events.jsonl")
}

func assertCompletedWorkdirCreatedEvent(t *testing.T, events []ir.CanonicalEvent) {
	t.Helper()

	assertRunEvent(t, events, ir.EventWorkdirCreated, func(ev ir.CanonicalEvent) {
		if ev.NativeConfigVersion != "" {
			t.Fatalf("WorkdirCreated must remain pre-execute without NCV: %+v", ev)
		}
		if ev.RiidoDaemonVersion != "riido-agentd v1.2.3" {
			t.Fatalf("daemon version not stamped: %+v", ev)
		}
	})
}
