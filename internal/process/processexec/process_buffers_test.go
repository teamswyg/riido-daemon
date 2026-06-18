package processexec

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestRealProcessDefaultBuffersMatchProviderRuntimeBackpressureSSOT(t *testing.T) {
	p := New()
	proc, err := p.Start(context.Background(), process.Command{
		Executable: "/bin/sh",
		Args:       []string{"-c", "printf ok"},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if got := cap(proc.Stdout()); got != process.DefaultStdoutBuffer {
		t.Fatalf("stdout buffer = %d, want %d", got, process.DefaultStdoutBuffer)
	}
	if got := cap(proc.Stderr()); got != process.DefaultStderrBuffer {
		t.Fatalf("stderr buffer = %d, want %d", got, process.DefaultStderrBuffer)
	}
	_ = drainAll(proc.Stdout(), 2*time.Second)
	requireExit(t, proc)
}
