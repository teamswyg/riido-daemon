package process

import (
	"context"
	"testing"
)

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
