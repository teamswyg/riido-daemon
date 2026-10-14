package session

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestSessionWithoutProtocolDriverKeepsLegacyPath(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	tracker := &trackingAdapter{}
	sess, err := Start(context.Background(), Config{
		TaskID:  "t-legacy",
		Adapter: tracker,
		Process: fake,
		Spawn:   process.Command{Executable: "fake"},
	})
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		sess.runningForTest().EmitStdout([]byte("x"))
		sess.runningForTest().EmitExit(0, nil)
	}()
	<-sess.Result()
	<-sess.Done()
	if tracker.TranslateCalls() == 0 {
		t.Fatal("Translate must be called on the legacy path")
	}
}
