package saasplane

import "testing"

func assertBindingCacheClaimWave(t *testing.T, fake *fakeAssignmentServer) {
	t.Helper()
	if got := fake.requestCount("/v1/daemon/agent-bindings"); got != 1 {
		t.Fatalf("agent-bindings request count = %d, want 1", got)
	}
	if got := len(fake.pollRequestsFor("agent-codex")); got != 1 {
		t.Fatalf("agent-codex poll count = %d, want 1", got)
	}
	if got := len(fake.pollRequestsFor("agent-cursor")); got != 1 {
		t.Fatalf("agent-cursor poll count = %d, want 1", got)
	}
}
