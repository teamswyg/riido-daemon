package runtimeactor

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestRuntimeActorSubmitAfterStop(t *testing.T) {
	a, _ := startActor(t, Config{
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true}},
		},
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = a.Stop(ctx)

	_, err := a.Submit(context.Background(), bridge.TaskRequest{ID: "t-late", Provider: "fake"})
	if !errors.Is(err, ErrActorStopped) {
		t.Fatalf("expected ErrActorStopped, got %v", err)
	}
}

func TestRuntimeActorTaskStatusIncluded(t *testing.T) {
	a, p := startActor(t, Config{
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true}},
		},
	})
	_, _ = a.Submit(context.Background(), bridge.TaskRequest{ID: "t-7", Provider: "fake"})
	_ = waitForRunning(t, p, 0, time.Second)

	s, _ := a.Status(context.Background())
	if len(s.RunningTasks) != 1 {
		t.Fatalf("RunningTasks: %+v", s.RunningTasks)
	}
	if s.RunningTasks[0].TaskID != "t-7" || s.RunningTasks[0].Provider != "fake" {
		t.Fatalf("RunningTasks entry: %+v", s.RunningTasks[0])
	}
	_ = strconv.Itoa // satisfy import if unused
}

func TestRuntimeActorStatusReplyWaitObservesStop(t *testing.T) {
	a := &Actor{
		cfg:       Config{RuntimeID: "rt-status-stop"},
		statusCh:  make(chan statusMsg, 1),
		stoppedCh: make(chan struct{}),
	}

	errCh := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		status, err := a.Status(ctx)
		if err != nil {
			errCh <- err
			return
		}
		if status.RuntimeID != "rt-status-stop" || status.Health != "stopped" {
			errCh <- errors.New("unexpected stopped status")
			return
		}
		errCh <- nil
	}()

	select {
	case <-a.statusCh:
	case <-time.After(time.Second):
		t.Fatal("Status did not enter reply wait")
	}
	close(a.stoppedCh)

	if err := <-errCh; err != nil {
		t.Fatalf("Status: %v", err)
	}
}

func TestRuntimeActorHeartbeatReplyWaitObservesStop(t *testing.T) {
	a := &Actor{
		cfg:       Config{RuntimeID: "rt-heartbeat-stop"},
		statusCh:  make(chan statusMsg, 1),
		stoppedCh: make(chan struct{}),
	}

	errCh := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		hb, err := a.HeartbeatPayload(ctx)
		if err != nil {
			errCh <- err
			return
		}
		if hb.RuntimeID != "rt-heartbeat-stop" {
			errCh <- errors.New("unexpected stopped heartbeat")
			return
		}
		errCh <- nil
	}()

	select {
	case <-a.statusCh:
	case <-time.After(time.Second):
		t.Fatal("HeartbeatPayload did not enter reply wait")
	}
	close(a.stoppedCh)

	if err := <-errCh; err != nil {
		t.Fatalf("HeartbeatPayload: %v", err)
	}
}
