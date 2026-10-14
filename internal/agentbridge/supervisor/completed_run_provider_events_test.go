package supervisor

import (
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
)

func assertCompletedProviderEvent(t *testing.T, events []ir.CanonicalEvent) {
	t.Helper()

	assertCompletedTextDeltaEvent(t, events)
	assertCompletedRunReportedDoneEvent(t, events)
}

func assertCompletedTextDeltaEvent(t *testing.T, events []ir.CanonicalEvent) {
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
}

func assertCompletedRunReportedDoneEvent(t *testing.T, events []ir.CanonicalEvent) {
	t.Helper()

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
