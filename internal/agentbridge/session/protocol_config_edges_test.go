package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestProtocolIONilProcessReportsStableError(t *testing.T) {
	var io *protocolIOImpl
	if err := io.WriteStdin(context.Background(), []byte("x")); !errors.Is(err, errProtocolIONil) {
		t.Fatalf("WriteStdin error=%v", err)
	}
	if err := io.CloseStdin(context.Background()); !errors.Is(err, errProtocolIONil) {
		t.Fatalf("CloseStdin error=%v", err)
	}
}

func TestProtocolIODelegatesWriteAndClose(t *testing.T) {
	running := process.NewFakeRunning()
	io := newProtocolIO(running)
	if err := io.WriteStdin(context.Background(), []byte("hello")); err != nil {
		t.Fatalf("WriteStdin: %v", err)
	}
	select {
	case got := <-running.StdinRecv():
		if string(got) != "hello" {
			t.Fatalf("stdin=%q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("stdin write was not delegated")
	}
	if err := io.CloseStdin(context.Background()); err != nil {
		t.Fatalf("CloseStdin: %v", err)
	}
	if _, ok := <-running.StdinRecv(); ok {
		t.Fatal("stdin channel should be closed")
	}
}

func TestValidateSessionConfigRequiresAdapterAndProcess(t *testing.T) {
	if err := validateSessionConfig(Config{Process: process.NewFake()}); err == nil {
		t.Fatal("expected missing adapter error")
	}
	if err := validateSessionConfig(Config{Adapter: &recordingAdapter{}}); err == nil {
		t.Fatal("expected missing process error")
	}
	if err := validateSessionConfig(Config{
		Adapter: &recordingAdapter{},
		Process: process.NewFake(),
	}); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}
}

func TestTerminateWithContextDefaultsEmptyStatusToFailed(t *testing.T) {
	started := startRecordingSession(t, "task-terminal-default", noEventRecordingAdapter(), nil)
	started.sess.TerminateWithContext(nil, agentbridge.Result{Error: "empty status"})
	res := waitResult(t, started.sess, time.Second)
	if res.Status != agentbridge.ResultFailed {
		t.Fatalf("status=%s, want failed", res.Status)
	}
	if res.Error != "empty status" {
		t.Fatalf("error=%q", res.Error)
	}
}
