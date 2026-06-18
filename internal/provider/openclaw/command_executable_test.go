package openclaw

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestBuildStartUsesRuntimeSelectedExecutable(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		Executable: "/opt/riido/bin/openclaw-supported",
		TaskID:     "task-openclaw-1",
		Prompt:     "do the thing",
	}, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Executable != "/opt/riido/bin/openclaw-supported" {
		t.Fatalf("executable = %q", cmd.Executable)
	}
}
