package supervisor

import (
	"strings"
	"testing"
	"time"
)

func assertRuntimeRegistrations(t *testing.T, source *runtimeRoutingSource) {
	t.Helper()
	seenRegistrations := map[string]bool{}
	for range 2 {
		select {
		case rt := <-source.registered:
			seenRegistrations[rt.RuntimeID] = true
			if rt.Provider != strings.TrimPrefix(rt.RuntimeID, "rt-") {
				t.Fatalf("provider-specific registration mismatch: %+v", rt)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("runtime registration was not published")
		}
	}
	if !seenRegistrations["rt-claude"] || !seenRegistrations["rt-codex"] {
		t.Fatalf("runtime registrations missing: %+v", seenRegistrations)
	}
}
