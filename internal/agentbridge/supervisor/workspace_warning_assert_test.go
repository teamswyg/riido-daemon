package supervisor

import (
	"strings"
	"testing"
	"time"
)

func expectWorkspaceWarning(t *testing.T, reporter *reporterProbe, want string) {
	t.Helper()
	select {
	case ev := <-reporter.events:
		if !strings.Contains(ev.Text, want) {
			t.Fatalf("workspace warning text = %q, want %q", ev.Text, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("workspace warning was not retried")
	}
}
