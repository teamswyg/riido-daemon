package session

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestSessionStdoutStderrBurstNoDeadlock(t *testing.T) {
	const burstChunks = 1024

	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	sess, err := Start(context.Background(), Config{
		TaskID:      "t-burst",
		RuntimeID:   "rt-1",
		Adapter:     &burstAdapter{},
		Process:     fake,
		Spawn:       process.Command{Executable: "burst"},
		EventBuffer: 8,
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	drained := make(chan int, 1)
	go countDrainedEvents(sess, drained)

	running := sess.runningForTest()
	go emitBurstStdout(running, burstChunks/2)
	go emitBurstStderr(running, burstChunks/2)

	select {
	case res := <-sess.Result():
		if res.Status != agentbridge.ResultCompleted {
			t.Fatalf("status: %s", res.Status)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("burst caused deadlock: Result never arrived")
	}

	running.EmitExit(0, nil)
	select {
	case n := <-drained:
		if n == 0 {
			t.Fatalf("no events drained")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("events channel never closed")
	}
}
