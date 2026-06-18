package main

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestNewDaemonRuntimeActorsUsesProviderSlotsForDynamicSaaSBindings(t *testing.T) {
	runtimes, err := newDaemonRuntimeActors(dynamicSaaSRuntimeSettings(), []agentbridge.Adapter{
		daemonTestAdapter{name: "codex"},
		daemonTestAdapter{name: "claude"},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(runtimes) != 2 {
		t.Fatalf("want one runtime per provider adapter, got %d", len(runtimes))
	}
	ctx := context.Background()
	want := map[string]string{"daemon-1:codex": "codex", "daemon-1:claude": "claude"}
	for _, rt := range runtimes {
		if err := rt.Start(ctx); err != nil {
			t.Fatalf("runtime start: %v", err)
		}
		t.Cleanup(func() { _ = rt.Stop(context.Background()) })
		status, err := rt.Status(ctx)
		if err != nil {
			t.Fatalf("status: %v", err)
		}
		provider, ok := want[status.RuntimeID]
		if !ok {
			t.Fatalf("unexpected runtime id %q", status.RuntimeID)
		}
		if len(status.Agents) != 0 {
			t.Fatalf("dynamic runtime %s should not use static agents: %+v", status.RuntimeID, status.Agents)
		}
		if len(status.Capabilities) != 1 || status.Capabilities[0].Provider != provider {
			t.Fatalf("runtime %s capabilities = %+v", status.RuntimeID, status.Capabilities)
		}
	}
}
