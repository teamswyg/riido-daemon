package runtimeactor

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func expectKillRequest(t *testing.T, running *blockingKillRunning, level lifecycle.ShutdownLevel, label string) {
	t.Helper()
	select {
	case <-running.KillRecv():
	case <-time.After(time.Second):
		t.Fatalf("%s did not request provider kill", label)
	}
	expectKillLevel(t, running, level, label)
}

func expectKillLevel(t *testing.T, running *blockingKillRunning, want lifecycle.ShutdownLevel, label string) {
	t.Helper()
	select {
	case level := <-running.KillLevelRecv():
		if level != want {
			t.Fatalf("%s kill level = %s, want %s", label, level, want)
		}
	case <-time.After(time.Second):
		t.Fatalf("%s did not carry provider kill level", label)
	}
}
