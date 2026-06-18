package supervisor

import (
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
)

func assertCompletedNativeConfigInjectedEvent(
	t *testing.T,
	events []ir.CanonicalEvent,
) {
	t.Helper()

	assertRunEvent(t, events, ir.EventNativeConfigInjected, func(ev ir.CanonicalEvent) {
		if ev.NativeConfigVersion == "" {
			t.Fatalf("NativeConfigInjected missing NCV: %+v", ev)
		}
	})
}

func assertCompletedWorkdirArchivedEvent(t *testing.T, events []ir.CanonicalEvent) {
	t.Helper()

	assertRunEvent(t, events, ir.EventWorkdirArchived, func(ev ir.CanonicalEvent) {
		if ev.NativeConfigVersion == "" {
			t.Fatalf("WorkdirArchived missing NCV: %+v", ev)
		}
	})
}
