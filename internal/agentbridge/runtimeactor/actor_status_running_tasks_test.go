package runtimeactor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

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
}
