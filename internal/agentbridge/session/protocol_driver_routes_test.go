package session

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

// --- A2. Raw events route through driver.OnRaw — adapter.Translate is
//     NOT called. ---

func TestSessionRoutesRawThroughProtocolDriver(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	tracker := &trackingAdapter{}
	driver := &fakeDriver{
		onRaw: func(_ context.Context, raw agentbridge.RawEvent, _ ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error) {
			if raw.Type != "chunk" {
				return nil, nil, nil
			}
			if string(raw.Bytes) == "DONE" {
				return []agentbridge.Event{{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "driver-done"}}}, nil, nil
			}
			return []agentbridge.Event{{Kind: agentbridge.EventTextDelta, Text: string(raw.Bytes)}}, nil, nil
		},
	}
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
	running := sess.runningForTest()
	go func() {
		running.EmitStdout([]byte("hello"))
		running.EmitStdout([]byte("DONE"))
		running.EmitExit(0, nil)
	}()

	select {
	case res := <-sess.Result():
		if res.Status != agentbridge.ResultCompleted || res.Output != "driver-done" {
			t.Fatalf("result: %+v", res)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no result")
	}

	// Wait for the actor's last write to be fully published before
	// reading the unsynchronized counter — Done() closes last.
	<-sess.Done()
	if tracker.TranslateCalls() != 0 {
		t.Fatalf("Translate must NOT be called when driver is installed, got %d calls", tracker.TranslateCalls())
	}
	if driver.rawCalls == 0 {
		t.Fatalf("OnRaw was never called")
	}
}

// --- A3. OnProcessExit is invoked and its events feed the reducer. ---

func TestSessionProtocolDriverProcessExitCleansUp(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	driver := &fakeDriver{
		onProcessExit: func(_ context.Context, status agentbridge.ProcessExitStatus, _ ProtocolIO) ([]agentbridge.Event, error) {
			// Simulate "fail every pending RPC request" — emit one
			// Error event the reducer will pass through (logged).
			return []agentbridge.Event{{
				Kind: agentbridge.EventError,
				Err:  "driver: 1 pending request cancelled due to process exit code " + itoa(status.Code),
			}}, nil
		},
	}
	sess, err := Start(context.Background(), Config{
		TaskID:         "t-exit",
		Adapter:        &trackingAdapter{},
		Process:        fake,
		Spawn:          process.Command{Executable: "fake"},
		ProtocolDriver: driver,
	})
	if err != nil {
		t.Fatal(err)
	}
	running := sess.runningForTest()

	// Drain events.
	gotDriverError := make(chan struct{}, 1)
	go func() {
		for ev := range sess.Events() {
			if ev.Kind == agentbridge.EventError && bytesContain([]byte(ev.Err), "driver:") {
				select {
				case gotDriverError <- struct{}{}:
				default:
				}
			}
		}
	}()

	running.EmitExit(2, nil)

	select {
	case <-gotDriverError:
	case <-time.After(2 * time.Second):
		t.Fatal("driver-emitted error event never reached events stream")
	}
	<-sess.Result()
	if driver.exitCalls != 1 {
		t.Fatalf("OnProcessExit calls: %d", driver.exitCalls)
	}
}

// --- A4. nil ProtocolDriver keeps the legacy Adapter.Translate path. ---

func TestSessionWithoutProtocolDriverKeepsLegacyPath(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	tracker := &trackingAdapter{}
	sess, err := Start(context.Background(), Config{
		TaskID:  "t-legacy",
		Adapter: tracker,
		Process: fake,
		Spawn:   process.Command{Executable: "fake"},
		// ProtocolDriver intentionally nil.
	})
	if err != nil {
		t.Fatal(err)
	}
	running := sess.runningForTest()
	go func() {
		running.EmitStdout([]byte("x"))
		running.EmitExit(0, nil)
	}()
	<-sess.Result()
	<-sess.Done()
	if tracker.TranslateCalls() == 0 {
		t.Fatal("Translate must be called on the legacy path")
	}
}
