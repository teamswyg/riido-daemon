package runtimeactor

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/process"
)

func expectFakeProcessKill(t *testing.T, running *process.FakeRunning, timeout string) {
	t.Helper()
	select {
	case <-running.KillRecv():
	case <-time.After(2 * time.Second):
		t.Fatal(timeout)
	}
}

func expectRunningSessions(t *testing.T, a *Actor, want int, timeout string) {
	t.Helper()
	end := time.Now().Add(2 * time.Second)
	for time.Now().Before(end) {
		status, _ := a.Status(t.Context())
		if status.RunningSessions == want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal(timeout)
}
