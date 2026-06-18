package runtimeactor

import (
	"context"
	"testing"
)

func actorStatusCapabilities(t *testing.T, actor *Actor) []Capability {
	t.Helper()
	status, err := actor.Status(context.Background())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	return status.Capabilities
}

func capabilitiesByProvider(caps []Capability) map[string]Capability {
	byProvider := make(map[string]Capability, len(caps))
	for _, capView := range caps {
		byProvider[capView.Provider] = capView
	}
	return byProvider
}
