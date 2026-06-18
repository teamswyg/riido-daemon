package runtimeactor

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestRuntimeActorStatusJSONShape(t *testing.T) {
	a, _ := startActor(t, Config{
		Owner:      "kim",
		DeviceName: "MacBook-Pro-SK.local",
		Agents: []AgentStatus{
			{AgentID: "riido", Name: "Riido", State: "online"},
		},
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true}},
		},
	})
	status, _ := a.Status(context.Background())

	if status.RuntimeID == "" {
		t.Fatal("RuntimeID empty")
	}
	if status.Health != "ok" {
		t.Fatalf("Health: %q", status.Health)
	}
	if status.StartedAt.IsZero() || status.MaxConcurrent == 0 {
		t.Fatalf("runtime liveness fields missing: %+v", status)
	}
	if status.Owner != "kim" || status.DeviceName != "MacBook-Pro-SK.local" {
		t.Fatalf("Figma runtime fields: owner=%q device=%q", status.Owner, status.DeviceName)
	}
	if len(status.Agents) != 1 || status.Agents[0].Name != "Riido" {
		t.Fatalf("Agents: %+v", status.Agents)
	}
}
