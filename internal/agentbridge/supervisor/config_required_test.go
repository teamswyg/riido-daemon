package supervisor

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestSupervisorRequiresRiidoDaemonVersion(t *testing.T) {
	rt := startRuntime(t, process.NewFake())
	_, err := New(Config{
		DaemonID: "daemon-1",
		Runtime:  rt,
		Source:   controlplane.NewMemorySource(),
		Reporter: newReporterProbe(),
	})
	if err == nil || !strings.Contains(err.Error(), "RiidoDaemonVersion is required") {
		t.Fatalf("New error = %v", err)
	}
}
