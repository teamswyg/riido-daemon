package session

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

type pollingFixtureDriver struct {
	fakeDriver
	polled  int
	cleaned bool
}

func (d *pollingFixtureDriver) PollInterval() time.Duration { return time.Millisecond }
func (d *pollingFixtureDriver) OnPoll(context.Context, ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error) {
	d.polled++
	return []agentbridge.Event{{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "external result"}}}, nil, nil
}

func (d *pollingFixtureDriver) OnTerminate(context.Context, ProtocolIO) (bool, error) {
	d.cleaned = true
	return false, nil
}

func TestSessionPollsProtocolThroughReducerAndCleansBeforeExit(t *testing.T) {
	proc := process.NewFake()
	proc.NextRunning = process.NewFakeRunning()
	driver := &pollingFixtureDriver{}
	sess, err := Start(t.Context(), Config{TaskID: "poll", RuntimeID: "runtime", Adapter: &trackingAdapter{}, Process: proc, Spawn: process.Command{Executable: "fake"}, ProtocolDriver: driver})
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for range sess.Events() {
		}
	}()
	select {
	case result := <-sess.Result():
		if result.Status != agentbridge.ResultCompleted || result.Output != "external result" {
			t.Fatal(result)
		}
	case <-time.After(time.Second):
		t.Fatal("protocol poll was not scheduled")
	}
	<-sess.Done()
	if driver.polled != 1 || !driver.cleaned {
		t.Fatalf("poll=%d cleanup=%v", driver.polled, driver.cleaned)
	}
}

type cleanupFixtureDriver struct {
	fakeDriver
	started      bool
	acknowledged bool
}

func (d *cleanupFixtureDriver) OnTerminate(ctx context.Context, io ProtocolIO) (bool, error) {
	if d.started {
		return false, nil
	}
	d.started = true
	return true, io.WriteStdin(ctx, []byte("delete-owned-entry"))
}

func (d *cleanupFixtureDriver) OnRaw(_ context.Context, raw agentbridge.RawEvent, _ ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error) {
	if string(raw.Bytes) == "deleted" {
		d.acknowledged = true
		return []agentbridge.Event{{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultCancelled}}}, nil, nil
	}
	return nil, nil, nil
}

func TestSessionCancellationDrainsQueueCleanupDespiteCancelledContext(t *testing.T) {
	proc := process.NewFake()
	proc.NextRunning = process.NewFakeRunning()
	driver := &cleanupFixtureDriver{}
	sess, err := Start(t.Context(), Config{TaskID: "cleanup", RuntimeID: "runtime", Adapter: &trackingAdapter{}, Process: proc, Spawn: process.Command{Executable: "fake"}, ProtocolDriver: driver})
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for range sess.Events() {
		}
	}()
	running := sess.runningForTest()
	sess.Cancel(nil)
	select {
	case written := <-running.StdinRecv():
		if string(written) != "delete-owned-entry" {
			t.Fatal(string(written))
		}
		running.EmitStdout([]byte("deleted"))
	case <-time.After(time.Second):
		t.Fatal("queue cleanup was skipped")
	}
	select {
	case result := <-sess.Result():
		if result.Status != agentbridge.ResultCancelled {
			t.Fatal(result)
		}
	case <-time.After(time.Second):
		t.Fatal("cancellation did not finish after acknowledgement")
	}
	<-sess.Done()
	if !driver.acknowledged {
		t.Fatal("transport killed before queue cleanup acknowledgement")
	}
}
