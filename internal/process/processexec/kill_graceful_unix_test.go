//go:build !windows

package processexec

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func TestGracefulKillUsesSIGTERMBeforeForcedTimeout(t *testing.T) {
	p := New()
	proc, err := p.Start(context.Background(), process.Command{
		Executable: "/bin/sh",
		Args: []string{
			"-c",
			"trap 'exit 0' TERM; sleep 30 & child=$!; wait $child",
		},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	ctx, cancel := lifecycle.DetachedShutdown(lifecycle.ShutdownGraceful, 2*time.Second)
	defer cancel()
	if err := proc.Kill(ctx.Context()); err != nil {
		t.Fatalf("Kill: %v", err)
	}
	assertGracefulKillStatus(t, proc)
}

func assertGracefulKillStatus(t *testing.T, proc process.RunningProcess) {
	t.Helper()
	select {
	case status := <-proc.Exited():
		if status.Code == 137 {
			t.Fatalf("expected graceful SIGTERM path before forced timeout, got SIGKILL status %+v", status)
		}
		if status.Code != 0 && status.Code != 143 {
			t.Fatalf("expected shell trap exit 0 or SIGTERM exit 143, got %+v", status)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("process group did not exit after graceful Kill")
	}
}
