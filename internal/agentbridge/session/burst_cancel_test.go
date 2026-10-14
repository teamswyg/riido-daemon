package session

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestSessionCancelDuringBurstDoesNotDeadlock(t *testing.T) {
	const burstChunks = 2048

	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	sess, _ := Start(context.Background(), Config{
		TaskID:      "t-burst-cancel",
		RuntimeID:   "rt-1",
		Adapter:     &burstAdapter{},
		Process:     fake,
		Spawn:       process.Command{Executable: "burst"},
		EventBuffer: 8,
	})
	running := sess.runningForTest()

	drained := make(chan struct{})
	go closeAfterEvents(sess, drained)

	burstDone := make(chan struct{})
	go emitCancelableBurst(running, burstChunks, burstDone)

	time.Sleep(20 * time.Millisecond)
	sess.Cancel(nil)

	select {
	case res := <-sess.Result():
		if res.Status != agentbridge.ResultCancelled {
			t.Fatalf("status: %s", res.Status)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("cancel-during-burst caused deadlock")
	}

	safeEmitExit(running, 0, nil)
	select {
	case <-drained:
	case <-time.After(2 * time.Second):
		t.Fatal("events channel didn't close after cancel")
	}
	<-burstDone
}
