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

// --- A5. OnStart returning error terminates the session with Failed. ---

func TestSessionProtocolDriverOnStartErrorTerminatesFailed(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	driver := &fakeDriver{
		onStart: func(_ context.Context, _ ProtocolIO) error {
			return errors.New("init handshake failed")
		},
	}
	sess, err := Start(context.Background(), Config{
		TaskID:         "t-onstart-err",
		Adapter:        &trackingAdapter{},
		Process:        fake,
		Spawn:          process.Command{Executable: "fake"},
		ProtocolDriver: driver,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Drain events so the actor doesn't block.
	go func() {
		for range sess.Events() {
		}
	}()

	select {
	case res := <-sess.Result():
		if res.Status != agentbridge.ResultFailed {
			t.Fatalf("status: %s", res.Status)
		}
		if res.Error == "" || !strings.Contains(res.Error, "init handshake failed") {
			t.Fatalf("result error: %q", res.Error)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("session never terminated on OnStart error")
	}

	// Kill the fake process if it's still alive so test cleanup doesn't leak.
	running := sess.runningForTest()
	func() {
		defer func() { _ = recover() }()
		running.EmitExit(0, nil)
	}()
}
