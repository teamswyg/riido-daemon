package supervisor

import (
	"testing"
	"time"
)

func expectStartReportAttempt(t *testing.T, reporter *startFailReporter, want string) {
	t.Helper()
	select {
	case got := <-reporter.attempted:
		if got != want {
			t.Fatalf("start report task = %q, want %q", got, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("start report was not attempted")
	}
}
