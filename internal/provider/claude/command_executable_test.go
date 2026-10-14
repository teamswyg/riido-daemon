package claude

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestBuildStartExecutableOverride(t *testing.T) {
	cmd, _ := BuildStart(agentbridge.StartRequest{}, StartOptions{
		Executable:     "/opt/anthropic/claude",
		PermissionMode: PermissionModeApproval,
	})
	if cmd.Executable != "/opt/anthropic/claude" {
		t.Fatalf("executable override lost: %q", cmd.Executable)
	}
}

func TestBuildStartUsesRuntimeSelectedExecutable(t *testing.T) {
	cmd, _ := BuildStart(agentbridge.StartRequest{
		Executable: "/opt/riido/bin/claude-selected",
	}, safeStartOptions())
	if cmd.Executable != "/opt/riido/bin/claude-selected" {
		t.Fatalf("runtime-selected executable lost: %q", cmd.Executable)
	}
}
