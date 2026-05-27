package process

import (
	"context"
	"testing"
)

func TestFakeProcessLifecycle(t *testing.T) {
	fake := NewFake()
	fake.NextRunning = NewFakeRunning()

	proc, err := fake.Start(context.Background(), Command{Executable: "x", Args: []string{"--flag"}})
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	running, ok := proc.(*FakeRunning)
	if !ok {
		t.Fatalf("expected *FakeRunning, got %T", proc)
	}
	if running.Command().Executable != "x" {
		t.Fatalf("command not stored: %+v", running.Command())
	}

	go func() {
		running.EmitStdout([]byte("line1\n"))
		running.EmitStderr([]byte("warn\n"))
		running.EmitExit(0, nil)
	}()

	var stdout []byte
	for chunk := range running.Stdout() {
		stdout = append(stdout, chunk...)
	}
	if string(stdout) != "line1\n" {
		t.Fatalf("stdout: got %q", stdout)
	}
	var stderr []byte
	for chunk := range running.Stderr() {
		stderr = append(stderr, chunk...)
	}
	if string(stderr) != "warn\n" {
		t.Fatalf("stderr: got %q", stderr)
	}

	status, ok := <-running.Exited()
	if !ok {
		t.Fatal("expected exit status before channel close")
	}
	if status.Code != 0 || status.Err != nil {
		t.Fatalf("unexpected exit status: %+v", status)
	}
	if _, ok := <-running.Exited(); ok {
		t.Fatal("expected Exited channel to be closed after delivering status")
	}
}

func TestFakeProcessDefaultBuffersMatchProviderRuntimeBackpressureSSOT(t *testing.T) {
	running := NewFakeRunning()
	if got := cap(running.Stdout()); got != DefaultStdoutBuffer {
		t.Fatalf("stdout buffer = %d, want %d", got, DefaultStdoutBuffer)
	}
	if got := cap(running.Stderr()); got != DefaultStderrBuffer {
		t.Fatalf("stderr buffer = %d, want %d", got, DefaultStderrBuffer)
	}
}

func TestFakeProcessWriteStdinAndKill(t *testing.T) {
	fake := NewFake()
	fake.NextRunning = NewFakeRunning()
	proc, _ := fake.Start(context.Background(), Command{Executable: "x"})
	running := proc.(*FakeRunning)

	if err := running.WriteStdin([]byte("hello\n")); err != nil {
		t.Fatalf("writestdin: %v", err)
	}
	select {
	case got := <-running.StdinRecv():
		if string(got) != "hello\n" {
			t.Fatalf("stdin: %q", got)
		}
	default:
		t.Fatal("expected stdin chunk")
	}

	if err := running.Kill(context.Background()); err != nil {
		t.Fatalf("kill: %v", err)
	}
	select {
	case <-running.KillRecv():
	default:
		t.Fatal("expected kill signal")
	}
}
