package session

import (
	"context"
	"errors"
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
		if res.Error == "" || !bytesContain([]byte(res.Error), "init handshake failed") {
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

// --- helpers ---

func bytesContain(b []byte, sub string) bool {
	if len(sub) > len(b) {
		return false
	}
	for i := 0; i+len(sub) <= len(b); i++ {
		if string(b[i:i+len(sub)]) == sub {
			return true
		}
	}
	return false
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
