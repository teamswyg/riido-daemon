package controlplane

import (
	"context"
	"testing"
)

func registerRuntimeAvailability(
	t *testing.T,
	src *FileQueueSource,
	runtimeID string,
	claude bool,
	codex bool,
) {
	t.Helper()

	if err := src.RegisterRuntime(context.Background(), RuntimeRegistration{
		DaemonID:  "daemon-" + runtimeID,
		RuntimeID: runtimeID,
		Provider:  "multi",
		Capabilities: map[string]bool{
			"provider.claude.available": claude,
			"provider.codex.available":  codex,
		},
	}); err != nil {
		t.Fatal(err)
	}
}
