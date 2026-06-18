package main

import "testing"

func assertDaemonRuntimeCapability(
	t *testing.T,
	rt daemonRuntimeStatus,
	c daemonCapabilityStatus,
	wantProviders map[string]bool,
	out string,
) {
	t.Helper()
	if rt.RuntimeID != "daemon-test-1:"+c.Provider {
		t.Fatalf("runtime_id/provider mismatch: runtime_id=%q provider=%q\n%s", rt.RuntimeID, c.Provider, out)
	}
	if _, ok := wantProviders[c.Provider]; ok {
		wantProviders[c.Provider] = true
	} else {
		t.Fatalf("unexpected provider capability: %+v\n%s", c, out)
	}
	if c.ProtocolKind == "" || c.AdapterID == "" || c.AdapterVersion == "" ||
		c.ProtocolVersion == "" || c.CompatibilityStatus == "" || c.CapabilityFingerprint == "" {
		t.Fatalf("capability missing C3 projection fields: %+v\n%s", c, out)
	}
}
