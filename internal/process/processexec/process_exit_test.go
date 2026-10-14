package processexec

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestRealNonZeroExit(t *testing.T) {
	p := New()
	proc, _ := p.Start(context.Background(), process.Command{
		Executable: "/bin/sh",
		Args:       []string{"-c", "exit 7"},
	})
	_ = drainAll(proc.Stdout(), time.Second)
	status := <-proc.Exited()
	if status.Code != 7 {
		t.Fatalf("exit code: %d", status.Code)
	}
}

func TestRealContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	p := New()
	proc, _ := p.Start(ctx, process.Command{
		Executable: "/bin/sh",
		Args:       []string{"-c", "sleep 30"},
	})
	cancel()
	select {
	case <-proc.Exited():
	case <-time.After(3 * time.Second):
		t.Fatal("context cancel didn't terminate process")
	}
}
