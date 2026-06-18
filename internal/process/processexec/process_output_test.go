package processexec

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestRealEcho(t *testing.T) {
	p := New()
	proc, err := p.Start(context.Background(), process.Command{
		Executable: "/bin/sh",
		Args:       []string{"-c", "echo hello"},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	stdout := drainAll(proc.Stdout(), 2*time.Second)
	if string(stdout) != "hello\n" {
		t.Fatalf("stdout: %q", stdout)
	}
	if status := requireExit(t, proc); status.Code != 0 {
		t.Fatalf("exit: %+v", status)
	}
}

func TestRealStderr(t *testing.T) {
	p := New()
	proc, _ := p.Start(context.Background(), process.Command{
		Executable: "/bin/sh",
		Args:       []string{"-c", "echo warn 1>&2"},
	})
	stderr := drainAll(proc.Stderr(), 2*time.Second)
	if string(stderr) != "warn\n" {
		t.Fatalf("stderr: %q", stderr)
	}
	<-proc.Exited()
}
