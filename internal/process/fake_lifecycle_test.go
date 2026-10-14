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

	assertFakeLifecycleOutput(t, running)
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

func assertFakeLifecycleOutput(t *testing.T, running *FakeRunning) {
	t.Helper()
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
}
