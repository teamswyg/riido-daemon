package runtimeactor

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

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

	claudeLike := &stubAdapter{name: "claude", detected: claudeCapabilityDetectResult(binary)}
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
