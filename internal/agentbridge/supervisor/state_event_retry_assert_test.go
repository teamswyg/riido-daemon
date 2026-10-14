package supervisor

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func expectStateEventAttempt(t *testing.T, reporter *stateEventRetryReporter, want int) {
	t.Helper()
	select {
	case got := <-reporter.attempts:
		if got != want {
			t.Fatalf("state event report attempt = %d, want %d", got, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("state event report attempt %d was not observed", want)
	}
}

func expectRetriedRunningEvent(t *testing.T, reporter *reporterProbe) {
	t.Helper()
	timer := time.After(2 * time.Second)
	for {
		select {
		case ev := <-reporter.events:
			if ev.Kind == agentbridge.EventLifecycle && ev.Phase == agentbridge.StateRunning {
				return
			}
		case <-timer:
			t.Fatal("retried running event was not reported")
		}
	}
}
