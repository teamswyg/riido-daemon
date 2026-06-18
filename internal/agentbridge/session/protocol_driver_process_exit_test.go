package session

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestSessionProtocolDriverProcessExitCleansUp(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	driver := &fakeDriver{onProcessExit: protocolDriverExitHandler}
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

	gotDriverError := watchDriverError(sess)
	sess.runningForTest().EmitExit(2, nil)

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

func protocolDriverExitHandler(
	_ context.Context,
	status agentbridge.ProcessExitStatus,
	_ ProtocolIO,
) ([]agentbridge.Event, error) {
	return []agentbridge.Event{{
		Kind: agentbridge.EventError,
		Err:  "driver: 1 pending request cancelled due to process exit code " + strconv.Itoa(status.Code),
	}}, nil
}

func watchDriverError(sess *Session) <-chan struct{} {
	gotDriverError := make(chan struct{}, 1)
	go func() {
		for ev := range sess.Events() {
			if ev.Kind == agentbridge.EventError && strings.Contains(ev.Err, "driver:") {
				gotDriverError <- struct{}{}
				return
			}
		}
	}()
	return gotDriverError
}
