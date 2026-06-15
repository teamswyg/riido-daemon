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
