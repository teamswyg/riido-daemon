package codex

import (
	"slices"
	"testing"
)

func TestBlockedArgsCoverProtocolCritical(t *testing.T) {
	for _, want := range []string{"--listen"} {
		if !slices.Contains(BlockedArgs(), want) {
			t.Fatalf("BlockedArgs missing %q: %v", want, BlockedArgs())
		}
	}
}

func TestUnsafeBypassArgsCoverSecuritySSOTSurfaces(t *testing.T) {
	for _, want := range []string{
		"--yolo",
		"--dangerously-bypass-approvals-and-sandbox",
	} {
		if !slices.Contains(UnsafeBypassArgs(), want) {
			t.Fatalf("UnsafeBypassArgs missing %q: %v", want, UnsafeBypassArgs())
		}
	}
}

func TestSandboxOverrideArgsCoverDaemonOwnedSandboxSelection(t *testing.T) {
	for _, want := range []string{"--sandbox", "-s"} {
		if !slices.Contains(SandboxOverrideArgs(), want) {
			t.Fatalf("SandboxOverrideArgs missing %s: %v", want, SandboxOverrideArgs())
		}
	}
}
