package runtimeactor

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	providercap "github.com/teamswyg/riido-contracts/provider/capability"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestRuntimeActorDetectedFingerprintHashesExecutable(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "claude")
	content := []byte("provider binary v1\n")
	if err := os.WriteFile(binary, content, 0o755); err != nil {
		t.Fatal(err)
	}
	wantSum := sha256.Sum256(content)
	want := hex.EncodeToString(wantSum[:])

	claudeLike := &stubAdapter{name: "claude", detected: agentbridge.DetectResult{
		Available:         true,
		Version:           "2.1.150",
		Executable:        binary,
		SupportsStreaming: true,
		SupportsResume:    true,
		SupportsSystem:    true,
		SupportsMaxTurns:  true,
		SupportsMCP:       true,
		SupportsToolHooks: true,
		SupportsUsage:     true,
	}}
	a, _ := startActor(t, Config{
		RuntimeID:           "rt-detected-fp",
		PolicyBundleVersion: "policy-bundle.test.v1",
		Adapters:            []agentbridge.Adapter{claudeLike},
	})

	status, err := a.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	capability := status.Capabilities[0]
	if capability.DetectedFingerprint != want {
		t.Fatalf("detected fingerprint = %q, want %q", capability.DetectedFingerprint, want)
	}
	withFingerprint := capability.CapabilityFingerprint
	if withFingerprint == "" {
		t.Fatalf("capability fingerprint missing: %+v", capability)
	}

	noBinary := *claudeLike
	noBinary.detected.Executable = "claude"
	a, _ = startActor(t, Config{
		RuntimeID:           "rt-detected-fp-empty",
		PolicyBundleVersion: "policy-bundle.test.v1",
		Adapters:            []agentbridge.Adapter{&noBinary},
	})
	status, err = a.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if status.Capabilities[0].DetectedFingerprint != "" {
		t.Fatalf("non-absolute executable must not be fingerprinted: %+v", status.Capabilities[0])
	}
	if status.Capabilities[0].CapabilityFingerprint == withFingerprint {
		t.Fatal("capability fingerprint must include detected fingerprint input")
	}
}

func TestRuntimeActorReconcilesUnavailableProviderAsBlocked(t *testing.T) {
	a, _ := startActor(t, Config{
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "cursor", detected: agentbridge.DetectResult{
				Available: false,
				Reason:    "cursor-agent missing",
			}},
		},
	})

	status, err := a.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	capability := status.Capabilities[0]
	if capability.Available {
		t.Fatalf("capability should be unavailable: %+v", capability)
	}
	if capability.CompatibilityStatus != string(providercap.CompatBlocked) {
		t.Fatalf("unavailable provider must be blocked: %+v", capability)
	}
	if capability.ProtocolKind != string(providercap.ProtocolCursorAgentStreamJSON) {
		t.Fatalf("cursor protocol kind missing: %+v", capability)
	}
}

func TestRuntimeActorCapabilityFingerprintIncludesPolicyBundle(t *testing.T) {
	detected := agentbridge.DetectResult{
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
	}
	start := func(policy string) string {
		t.Helper()
		a, _ := startActor(t, Config{
			RuntimeID:           "rt-policy",
			PolicyBundleVersion: policy,
			Adapters:            []agentbridge.Adapter{&stubAdapter{name: "claude", detected: detected}},
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

// --- 2. daemon status populates runtimes from Actor ---
// This is exercised in cmd/riido/daemon_test.go after wiring; here we
// verify the JSON shape is producible from Actor.Status alone.

func TestRuntimeActorStatusJSONShape(t *testing.T) {
	a, _ := startActor(t, Config{
		Owner:      "kim",
		DeviceName: "MacBook-Pro-SK.local",
		Agents: []AgentStatus{
			{AgentID: "riido", Name: "Riido", State: "online"},
		},
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true}},
		},
	})
	status, _ := a.Status(context.Background())

	// Required runtime status fields per provider-runtime SSOT.
	if status.RuntimeID == "" {
		t.Fatal("RuntimeID empty")
	}
	if status.Health != "ok" {
		t.Fatalf("Health: %q", status.Health)
	}
	if status.StartedAt.IsZero() {
		t.Fatal("StartedAt zero")
	}
	if status.MaxConcurrent == 0 {
		t.Fatal("MaxConcurrent zero")
	}
	if status.Owner != "kim" || status.DeviceName != "MacBook-Pro-SK.local" {
		t.Fatalf("Figma runtime fields: owner=%q device=%q", status.Owner, status.DeviceName)
	}
	if len(status.Agents) != 1 || status.Agents[0].Name != "Riido" {
		t.Fatalf("Agents: %+v", status.Agents)
	}
}
