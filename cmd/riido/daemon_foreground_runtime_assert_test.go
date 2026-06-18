package main

import "testing"

func assertDaemonRuntimes(t *testing.T, status daemonStatusJSON, out string) {
	t.Helper()
	if len(status.Runtimes) != 4 {
		t.Fatalf("want 4 provider runtimes, got %d: %+v\n%s", len(status.Runtimes), status.Runtimes, out)
	}
	wantProviders := map[string]bool{"claude": false, "codex": false, "openclaw": false, "cursor": false}
	for _, rt := range status.Runtimes {
		assertDaemonRuntime(t, rt, wantProviders, out)
	}
	for provider, seen := range wantProviders {
		if !seen {
			t.Fatalf("capability missing for provider %s", provider)
		}
	}
}

func assertDaemonRuntime(
	t *testing.T,
	rt daemonRuntimeStatus,
	wantProviders map[string]bool,
	out string,
) {
	t.Helper()
	if rt.Health != "ok" {
		t.Fatalf("runtime health: %q", rt.Health)
	}
	if rt.Owner != "kim" || rt.DeviceName != "MacBook-Pro-SK.local" {
		t.Fatalf("runtime UI fields missing: owner=%q device=%q\n%s", rt.Owner, rt.DeviceName, out)
	}
	if len(rt.Agents) != 2 || rt.Agents[0].Name != "Riido" || rt.Agents[1].Name != "Orion" {
		t.Fatalf("runtime agents mismatch: %+v\n%s", rt.Agents, out)
	}
	if len(rt.Capabilities) != 1 {
		t.Fatalf("provider runtime should expose exactly one capability, got %d: %+v", len(rt.Capabilities), rt.Capabilities)
	}
	assertDaemonRuntimeCapability(t, rt, rt.Capabilities[0], wantProviders, out)
}
