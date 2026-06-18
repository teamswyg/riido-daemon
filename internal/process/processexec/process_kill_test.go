package processexec

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestRealKill(t *testing.T) {
	p := New()
	proc, _ := p.Start(context.Background(), process.Command{
		Executable: "/bin/sh",
		Args:       []string{"-c", "sleep 30"},
	})
	if err := proc.Kill(context.Background()); err != nil {
		t.Fatalf("Kill: %v", err)
	}
	select {
	case status := <-proc.Exited():
		if status.Code == 0 {
			t.Fatalf("expected non-zero exit after kill, got %+v", status)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("process did not exit after Kill")
	}
}
