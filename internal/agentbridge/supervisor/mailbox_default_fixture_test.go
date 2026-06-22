package supervisor

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/process"
)

func newDefaultMailboxSupervisor(t *testing.T) *Actor {
	t.Helper()
	rt := startRuntime(t, process.NewFake())
	actor, err := New(Config{
		DaemonID:           "daemon-mailbox-default",
		RiidoDaemonVersion: testRiidoDaemonVersion,
		Runtime:            rt,
		Source:             newRuntimeRoutingSource(nil),
		Reporter:           newReporterProbe(),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return actor
}
