//go:build !windows

package processexec

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestKillTerminatesProcessGroupAndClosesPipes(t *testing.T) {
	p := New()
	proc, err := p.Start(context.Background(), process.Command{
		Executable: "/bin/sh",
		Args: []string{
			"-c",
			"sleep 30 & child=$!; wait $child",
		},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := proc.Kill(context.Background()); err != nil {
		t.Fatalf("Kill: %v", err)
	}
	assertProcessGroupKilled(t, proc)
	assertStdoutClosed(t, proc)
}

func assertProcessGroupKilled(t *testing.T, proc process.RunningProcess) {
	t.Helper()
	select {
	case status := <-proc.Exited():
		if status.Code == 0 {
			t.Fatalf("expected non-zero exit after process-group kill, got %+v", status)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("process group did not exit after Kill")
	}
}

func assertStdoutClosed(t *testing.T, proc process.RunningProcess) {
	t.Helper()
	select {
	case _, ok := <-proc.Stdout():
		if ok {
			t.Fatal("stdout channel still open after process-group kill")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("stdout pipe stayed open; background child may have survived")
	}
}
