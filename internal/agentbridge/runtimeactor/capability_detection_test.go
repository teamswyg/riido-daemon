package runtimeactor

import (
	"context"
	"errors"
	"testing"
	"time"

	providercap "github.com/teamswyg/riido-contracts/provider/capability"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func startActor(t *testing.T, cfg Config) (*Actor, *fakeProcess) {
	t.Helper()
	if cfg.Process == nil {
		cfg.Process = newFakeProcess()
	}
	if cfg.MaxConcurrent == 0 {
		cfg.MaxConcurrent = 2
	}
	if cfg.MailboxSize == 0 {
		cfg.MailboxSize = 8
	}
	if cfg.RuntimeID == "" {
		cfg.RuntimeID = "rt-test"
	}
	a, err := New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := a.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = a.Stop(ctx)
	})
	return a, cfg.Process.(*fakeProcess)
}

// --- 1. Detects capabilities on start ---

func TestRuntimeActorDetectsCapabilitiesOnStart(t *testing.T) {
	avail := &stubAdapter{name: "available", detected: agentbridge.DetectResult{Available: true, Version: "1.0", Executable: "/usr/bin/available"}}
	missing := &stubAdapter{name: "missing", detected: agentbridge.DetectResult{Available: false, Reason: "not installed"}}

	a, _ := startActor(t, Config{
		Adapters: []agentbridge.Adapter{avail, missing},
	})

	status, err := a.Status(context.Background())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if len(status.Capabilities) != 2 {
		t.Fatalf("want 2 capabilities, got %d: %+v", len(status.Capabilities), status.Capabilities)
	}
	names := map[string]Capability{}
	for _, c := range status.Capabilities {
		names[c.Provider] = c
	}
	if !names["available"].Available || names["available"].Version != "1.0" {
		t.Fatalf("available capability: %+v", names["available"])
	}
	if names["missing"].Available || names["missing"].Reason != "not installed" {
		t.Fatalf("missing capability: %+v", names["missing"])
	}
}

func TestRuntimeActorRefreshesUnavailableCapabilityAfterTTL(t *testing.T) {
	now := time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC)
	provider := &stubAdapter{name: "late", detected: agentbridge.DetectResult{Available: false, Reason: "not installed"}}

	a, p := startActor(t, Config{
		Adapters:               []agentbridge.Adapter{provider},
		CapabilityRefreshEvery: time.Second,
		Now:                    func() time.Time { return now },
	})

	provider.detected = agentbridge.DetectResult{Available: true, Version: "1.2.3", Executable: "/usr/local/bin/late"}
	_, err := a.Submit(context.Background(), bridge.TaskRequest{ID: "before-ttl", Provider: "late"})
	if !errors.Is(err, ErrProviderUnavailable) {
		t.Fatalf("submit before ttl should use cached unavailable capability, got %v", err)
	}
	if p.count() != 0 {
		t.Fatalf("submit before ttl should not spawn provider, got %d", p.count())
	}

	now = now.Add(2 * time.Second)
	if _, err := a.Submit(context.Background(), bridge.TaskRequest{ID: "after-ttl", Provider: "late"}); err != nil {
		t.Fatalf("submit after ttl should refresh provider capability: %v", err)
	}
	if p.count() != 1 {
		t.Fatalf("submit after ttl should spawn provider once, got %d", p.count())
	}

	status, err := a.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(status.Capabilities) != 1 || !status.Capabilities[0].Available || status.Capabilities[0].Version != "1.2.3" {
		t.Fatalf("refreshed capability not projected to status: %+v", status.Capabilities)
	}
}

func TestRuntimeActorReconcilesDetectResultToProviderCapability(t *testing.T) {
	fixedNow := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	claudeLike := &stubAdapter{name: "claude", detected: agentbridge.DetectResult{
		Available:         true,
		Version:           "2.1.150",
		Executable:        "/usr/local/bin/claude",
		SupportsStreaming: true,
		SupportsResume:    true,
		SupportsSystem:    true,
		SupportsMaxTurns:  true,
		SupportsMCP:       true,
		SupportsToolHooks: true,
		SupportsUsage:     true,
	}}
	a, _ := startActor(t, Config{
		RuntimeID:           "rt-cap",
		PolicyBundleVersion: "policy-bundle.test.v1",
		Adapters:            []agentbridge.Adapter{claudeLike},
		Now:                 func() time.Time { return fixedNow },
	})

	status, err := a.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(status.Capabilities) != 1 {
		t.Fatalf("capabilities: %+v", status.Capabilities)
	}
	capability := status.Capabilities[0]
	if capability.ProtocolKind != string(providercap.ProtocolClaudeStreamJSON) {
		t.Fatalf("protocol kind: %+v", capability)
	}
	if capability.AdapterID != "claude" || capability.AdapterVersion != "riido-agentbridge-adapter.v1" || capability.ProtocolVersion != "v1" {
		t.Fatalf("execution fingerprint fields missing: %+v", capability)
	}
	if capability.CompatibilityStatus != string(providercap.CompatSupported) {
		t.Fatalf("compatibility status: %+v", capability)
	}
	if capability.CapabilityFingerprint == "" {
		t.Fatalf("fingerprint missing: %+v", capability)
	}
	if !capability.SupportsStreaming || !capability.SupportsResume || !capability.SupportsSystem ||
		!capability.SupportsMaxTurns || !capability.SupportsMCP || !capability.SupportsToolHooks ||
		!capability.SupportsUsage || !capability.SupportsWorktree {
		t.Fatalf("surface flags were not preserved: %+v", capability)
	}
	if capability.SupportsFileEvents {
		t.Fatalf("file events must stay false until a provider emits structured file events: %+v", capability)
	}
}

func TestRuntimeActorLeavesOpenClawWorktreeUnsupported(t *testing.T) {
	openclawLike := &stubAdapter{name: "openclaw", detected: agentbridge.DetectResult{
		Available:         true,
		Version:           "2026.5.22",
		Executable:        "/usr/local/bin/openclaw",
		SupportsStreaming: true,
		SupportsResume:    true,
		SupportsUsage:     true,
	}}
	a, _ := startActor(t, Config{
		RuntimeID: "rt-openclaw",
		Adapters:  []agentbridge.Adapter{openclawLike},
	})

	status, err := a.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	capability := status.Capabilities[0]
	if capability.SupportsWorktree {
		t.Fatalf("OpenClaw must not advertise daemon-selected worktree support without a native workspace surface: %+v", capability)
	}
}
