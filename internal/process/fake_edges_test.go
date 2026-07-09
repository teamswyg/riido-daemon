package process

import (
	"context"
	"errors"
	"testing"
)

func TestFakeRunningStartedRecvReceivesCommand(t *testing.T) {
	fake := NewFake()
	proc, err := fake.Start(context.Background(), Command{Executable: "riido", Args: []string{"daemon"}})
	if err != nil {
		t.Fatal(err)
	}
	running := proc.(*FakeRunning)
	select {
	case got := <-running.StartedRecv():
		if got.Executable != "riido" || got.Args[0] != "daemon" {
			t.Fatalf("unexpected started command: %+v", got)
		}
	default:
		t.Fatal("expected started command")
	}
}

func TestFakeRunningStdinBufferFullAndClose(t *testing.T) {
	running := NewFakeRunning()
	for i := 0; i < cap(running.stdin); i++ {
		if err := running.WriteStdin([]byte("x")); err != nil {
			t.Fatalf("write %d: %v", i, err)
		}
	}
	if err := running.WriteStdin([]byte("overflow")); err == nil {
		t.Fatal("expected stdin buffer full error")
	}
	if err := running.CloseStdin(); err != nil {
		t.Fatal(err)
	}
	if err := running.CloseStdin(); err != nil {
		t.Fatal("double close should be tolerated")
	}
}

func TestFakeRunningIgnoresEmitsAfterExit(t *testing.T) {
	running := NewFakeRunning()
	running.EmitExit(2, errors.New("done"))
	status := <-running.Exited()
	if status.Code != 2 || status.Err == nil {
		t.Fatalf("unexpected exit: %+v", status)
	}
	running.EmitStdout([]byte("late"))
	running.EmitStderr([]byte("late"))
	running.EmitExit(3, nil)
	if _, ok := <-running.Stdout(); ok {
		t.Fatal("stdout should stay closed")
	}
	if _, ok := <-running.Stderr(); ok {
		t.Fatal("stderr should stay closed")
	}
}
