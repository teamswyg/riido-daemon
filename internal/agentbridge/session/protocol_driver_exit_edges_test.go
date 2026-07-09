package session

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestProtocolIOErrorMessageStable(t *testing.T) {
	got := errProtocolIONil.Error()
	if got != "session: protocol io has no process" {
		t.Fatalf("error message=%q", got)
	}
}

func TestSessionProtocolDriverProcessExitErrorIsEmitted(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	driver := &fakeDriver{onProcessExit: func(
		_ context.Context,
		status agentbridge.ProcessExitStatus,
		_ ProtocolIO,
	) ([]agentbridge.Event, error) {
		if status.Code != 7 || !strings.Contains(status.Err, "boom") {
			t.Fatalf("status=%+v", status)
		}
		return []agentbridge.Event{{
			Kind:   agentbridge.EventResult,
			Result: agentbridge.Result{Status: agentbridge.ResultCompleted},
		}}, errors.New("driver exit failed")
	}}
	sess, err := Start(context.Background(), Config{
		TaskID:         "t-exit-driver-error",
		Adapter:        &trackingAdapter{},
		Process:        fake,
		Spawn:          process.Command{Executable: "fake"},
		ProtocolDriver: driver,
	})
	if err != nil {
		t.Fatal(err)
	}
	errCh := watchErrorText(sess)
	sess.runningForTest().EmitExit(7, errors.New("boom"))
	select {
	case got := <-errCh:
		if !strings.Contains(got, "driver exit failed") {
			t.Fatalf("error event=%q", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("driver process-exit error was not emitted")
	}
	res := waitResult(t, sess, 2*time.Second)
	if res.Status != agentbridge.ResultCompleted {
		t.Fatalf("result=%+v", res)
	}
}

func watchErrorText(sess *Session) <-chan string {
	errCh := make(chan string, 1)
	go func() {
		for ev := range sess.Events() {
			if ev.Kind == agentbridge.EventError {
				errCh <- ev.Err
				return
			}
		}
	}()
	return errCh
}
