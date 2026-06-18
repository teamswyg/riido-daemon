package runtimeactor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestRuntimeActorHeartbeatPayload(t *testing.T) {
	a, p := startActor(t, Config{
		RuntimeID:  "rt-42",
		DeviceName: "device-a",
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true, Version: "1.0"}},
		},
		MaxConcurrent: 3,
	})
	_, _ = a.Submit(context.Background(), bridge.TaskRequest{ID: "t-1", Provider: "fake"})
	_ = waitForRunning(t, p, 0, time.Second)

	hb, err := a.HeartbeatPayload(context.Background())
	if err != nil {
		t.Fatalf("HeartbeatPayload: %v", err)
	}
	if hb.RuntimeID != "rt-42" {
		t.Fatalf("id: %q", hb.RuntimeID)
	}
	if hb.DeviceName != "device-a" {
		t.Fatalf("device name: %q", hb.DeviceName)
	}
	if hb.SlotLimit != 3 || hb.SlotsInUse != 1 {
		t.Fatalf("slots: %+v", hb)
	}
	if len(hb.RunningTaskIDs) != 1 || hb.RunningTaskIDs[0] != "t-1" {
		t.Fatalf("running ids: %v", hb.RunningTaskIDs)
	}
}
