package runtimeactor

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestRuntimeActorCapabilityFingerprintIncludesPolicyBundle(t *testing.T) {
	start := func(policy string) string {
		t.Helper()
		a, _ := startActor(t, Config{
			RuntimeID:           "rt-policy",
			PolicyBundleVersion: policy,
			Adapters: []agentbridge.Adapter{
				&stubAdapter{name: "claude", detected: claudeCapabilityDetectResult("/usr/local/bin/claude")},
			},
		})
		status, err := a.Status(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		return status.Capabilities[0].CapabilityFingerprint
	}

	v1 := start("policy-bundle.test.v1")
	v2 := start("policy-bundle.test.v2")
	if v1 == "" || v2 == "" {
		t.Fatalf("fingerprint missing: v1=%q v2=%q", v1, v2)
	}
	if v1 == v2 {
		t.Fatal("capability fingerprint must change when policy bundle version changes")
	}
}
