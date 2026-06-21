package supervisor

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func expectTerminalAttempt(t *testing.T, reporter *terminalRetryReporter, want int) {
	t.Helper()
	select {
	case got := <-reporter.attempts:
		if got != want {
			t.Fatalf("terminal report attempt = %d, want %d", got, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("terminal report attempt %d was not observed", want)
	}
}

func expectTaskResultCompleted(t *testing.T, reporter *reporterProbe) {
	t.Helper()
	select {
	case res := <-reporter.results:
		if res.Status != agentbridge.ResultCompleted {
			t.Fatalf("terminal status = %s, want %s", res.Status, agentbridge.ResultCompleted)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("terminal result was not reported")
	}
}
