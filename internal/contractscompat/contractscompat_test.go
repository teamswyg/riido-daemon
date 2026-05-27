package contractscompat

import (
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/provider/capability"
	"github.com/teamswyg/riido-contracts/task"
)

func TestContractsBaseline(t *testing.T) {
	if !ir.EventTaskQueued.IsTransition() {
		t.Fatal("TaskQueued must remain a transition event")
	}
	if task.FSMSchemaVersion != 1 {
		t.Fatalf("FSMSchemaVersion = %d", task.FSMSchemaVersion)
	}
	if !task.ValidateTransition(task.StateCreated, task.StateQueued, ir.EventTaskQueued) {
		t.Fatal("Created -> Queued must remain legal")
	}

	fingerprint, err := capability.ComputeCapabilityFingerprint(capability.CapabilityFingerprintInput{
		ProviderKind:          capability.ProviderKind("codex"),
		ProtocolKind:          capability.ProtocolCodexAppServer,
		ProviderVersion:       "codex test",
		DetectedFingerprint:   capability.DetectedFingerprint("detected"),
		AdapterID:             "codex",
		AdapterVersion:        "riido-agentbridge-adapter.v1",
		ProtocolVersion:       "v1",
		DefaultSandboxMode:    "workspace-write",
		DefaultApprovalPolicy: "on-request",
		PolicyBundleVersion:   "policy-bundle.test.v1",
		ImportantSurfaceFlags: map[string]any{"structuredEventStream": true},
	})
	if err != nil {
		t.Fatalf("ComputeCapabilityFingerprint: %v", err)
	}
	if fingerprint == "" {
		t.Fatal("CapabilityFingerprint is empty")
	}
}
