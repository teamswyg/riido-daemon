package contractscompat

import (
	"testing"

	"github.com/teamswyg/riido-contracts/metadatakeys"
	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
	contractrunstate "github.com/teamswyg/riido-contracts/runstate"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/supervisor"
)

func TestDaemonRunStateConsumesContractsVocabulary(t *testing.T) {
	if agentbridge.StateRunning != contractrunstate.StateRunning {
		t.Fatalf("daemon running state = %q, want contracts %q", agentbridge.StateRunning, contractrunstate.StateRunning)
	}
	for _, state := range agentbridge.AllStates() {
		if state.Code() != contractrunstate.ParseRunStateCode(string(state)) {
			t.Fatalf("daemon state %q diverged from contracts runstate parser", state)
		}
	}
}

func TestDaemonMetadataKeysConsumeContractsVocabulary(t *testing.T) {
	tests := map[string]string{
		"supervisor workspace":  supervisor.MetadataWorkspaceID,
		"supervisor run":        supervisor.MetadataRunID,
		"supervisor native cfg": supervisor.MetadataNativeConfig,
		"control task":          controlplane.MetadataTaskID,
		"control lease":         controlplane.MetadataRuntimeLeaseID,
		"progress code":         agentbridge.ProgressMessageMetadataCode,
	}
	want := map[string]string{
		"supervisor workspace":  metadatakeys.WorkspaceID.String(),
		"supervisor run":        metadatakeys.RunID.String(),
		"supervisor native cfg": metadatakeys.NativeConfigDir.String(),
		"control task":          metadatakeys.TaskID.String(),
		"control lease":         metadatakeys.RuntimeLeaseID.String(),
		"progress code":         metadatakeys.ProgressMessageCode.String(),
	}
	for name, got := range tests {
		if got != want[name] {
			t.Fatalf("%s metadata key = %q, want %q", name, got, want[name])
		}
	}
}

func TestDaemonProviderCatalogConsumesContractsDefaults(t *testing.T) {
	for _, tt := range []struct {
		provider string
		modelID  string
	}{
		{provider: "codex", modelID: "codex-default"},
		{provider: "claude", modelID: "claude-default"},
		{provider: "claude_code", modelID: "claude-default"},
		{provider: "openclaw", modelID: "openclaw-default"},
		{provider: "cursor", modelID: "cursor-auto"},
		{provider: "other", modelID: "runtime-default"},
	} {
		if got := providercatalog.ModelOverride(tt.provider, tt.modelID); got != "" {
			t.Fatalf("ModelOverride(%q, %q) = %q, want empty", tt.provider, tt.modelID, got)
		}
	}
}
