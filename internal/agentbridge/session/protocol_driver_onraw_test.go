package session

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestSessionRoutesRawThroughProtocolDriver(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	tracker := &trackingAdapter{}
	driver := &fakeDriver{onRaw: protocolDriverRawHandler}
	sess, err := Start(context.Background(), Config{
		TaskID:         "t-onraw",
		Adapter:        tracker,
		Process:        fake,
		Spawn:          process.Command{Executable: "fake"},
		ProtocolDriver: driver,
	})
	if err != nil {
		t.Fatal(err)
	}
	emitProtocolDriverRawInput(sess.runningForTest())

	select {
	case res := <-sess.Result():
		if res.Status != agentbridge.ResultCompleted || res.Output != "driver-done" {
			t.Fatalf("result: %+v", res)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no result")
	}
	<-sess.Done()
	if tracker.TranslateCalls() != 0 {
		t.Fatalf("Translate must NOT be called when driver is installed, got %d calls", tracker.TranslateCalls())
	}
	if driver.rawCalls == 0 {
		t.Fatalf("OnRaw was never called")
	}
}

func protocolDriverRawHandler(
	_ context.Context,
	raw agentbridge.RawEvent,
	_ ProtocolIO,
) ([]agentbridge.Event, []agentbridge.Command, error) {
	if raw.Type != "chunk" {
		return nil, nil, nil
	}
	if string(raw.Bytes) == "DONE" {
		result := agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "driver-done"}
		return []agentbridge.Event{{Kind: agentbridge.EventResult, Result: result}}, nil, nil
	}
	return []agentbridge.Event{{Kind: agentbridge.EventTextDelta, Text: string(raw.Bytes)}}, nil, nil
}

type protocolDriverRawEmitter interface {
	EmitStdout([]byte)
	EmitExit(int, error)
}

func emitProtocolDriverRawInput(running protocolDriverRawEmitter) {
	go func() {
		running.EmitStdout([]byte("hello"))
		running.EmitStdout([]byte("DONE"))
		running.EmitExit(0, nil)
	}()
}
