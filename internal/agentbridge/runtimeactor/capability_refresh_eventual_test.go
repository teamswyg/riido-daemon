package runtimeactor

import (
	"testing"
	"time"
)

func awaitCapability(t *testing.T, actor *Actor, match func(Capability) bool) Capability {
	t.Helper()
	deadline := time.NewTimer(time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(time.Millisecond)
	defer tick.Stop()
	for {
		caps := actorStatusCapabilities(t, actor)
		for _, capability := range caps {
			if match(capability) {
				return capability
			}
		}
		select {
		case <-tick.C:
		case <-deadline.C:
			t.Fatalf("capability observation did not converge: %+v", caps)
		}
	}
}
