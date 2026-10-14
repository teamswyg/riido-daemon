package codex

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestBuildStartExecutableOverride(t *testing.T) {
	cmd, _ := BuildStart(agentbridge.StartRequest{}, StartOptions{Executable: "/opt/codex/bin/codex"})
	if cmd.Executable != "/opt/codex/bin/codex" {
		t.Fatalf("exe override lost: %q", cmd.Executable)
	}
}

func TestBuildStartUsesRuntimeSelectedExecutable(t *testing.T) {
	cmd, _ := BuildStart(agentbridge.StartRequest{Executable: "/opt/riido/bin/codex-selected"}, StartOptions{})
	if cmd.Executable != "/opt/riido/bin/codex-selected" {
		t.Fatalf("runtime-selected executable lost: %q", cmd.Executable)
	}
}
