package runtimeactor

import (
	"context"
	"testing"
	"time"

	providercap "github.com/teamswyg/riido-contracts/provider/capability"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
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
