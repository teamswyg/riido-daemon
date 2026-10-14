package session

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestSessionCallsProtocolDriverOnStart(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	driver := &fakeDriver{startStdin: []byte(`{"jsonrpc":"2.0","id":1,"method":"initialize"}` + "\n")}
	sess, err := Start(context.Background(), Config{
		TaskID:         "t-onstart",
		RuntimeID:      "rt-1",
		Adapter:        &trackingAdapter{},
		Process:        fake,
		Spawn:          process.Command{Executable: "fake"},
		ProtocolDriver: driver,
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	running := sess.runningForTest()

	select {
	case b := <-running.StdinRecv():
		if !strings.Contains(string(b), "initialize") {
			t.Fatalf("OnStart stdin: %q", b)
		}
	case <-time.After(time.Second):
		t.Fatal("OnStart stdin frame never written")
	}

	go func() {
		for range sess.Events() {
		}
	}()
	running.EmitExit(0, nil)
	<-sess.Result()
	<-sess.Done()
	if driver.startCalls != 1 {
		t.Fatalf("OnStart calls: %d", driver.startCalls)
	}
	if driver.closeCalls != 1 {
		t.Fatalf("OnClose calls: %d", driver.closeCalls)
	}
}
