package processexec

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/process"
)

func drainAll(ch <-chan []byte, deadline time.Duration) []byte {
	var out []byte
	timer := time.NewTimer(deadline)
	defer timer.Stop()
	for {
		select {
		case chunk, ok := <-ch:
			if !ok {
				return out
			}
			out = append(out, chunk...)
		case <-timer.C:
			return out
		}
	}
}

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

	select {
	case status := <-proc.Exited():
		if status.Code != 0 {
			t.Fatalf("exit: %+v", status)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no exit signal")
	}
}

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

func TestRealStdinPipe(t *testing.T) {
	p := New()
	proc, _ := p.Start(context.Background(), process.Command{
		Executable: "/bin/cat",
	})
	if err := proc.WriteStdin([]byte("hi\n")); err != nil {
		t.Fatalf("WriteStdin: %v", err)
	}
	if err := proc.CloseStdin(); err != nil {
		t.Fatalf("CloseStdin: %v", err)
	}
	stdout := drainAll(proc.Stdout(), 2*time.Second)
	if string(stdout) != "hi\n" {
		t.Fatalf("stdout: %q", stdout)
	}
	<-proc.Exited()
}

func TestEnvOverridesPreserveParentEnvironment(t *testing.T) {
	p := New()
	proc, err := p.Start(context.Background(), process.Command{
		Executable: "/bin/sh",
		Args:       []string{"-c", `test -n "$PATH" && printf "%s" "$RIIDO_TEST_ENV"`},
		Env:        []string{"RIIDO_TEST_ENV=ok"},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	stdout := drainAll(proc.Stdout(), 2*time.Second)
	if string(stdout) != "ok" {
		t.Fatalf("stdout: %q", stdout)
	}
	status := <-proc.Exited()
	if status.Code != 0 {
		t.Fatalf("exit: %+v", status)
	}
}

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
