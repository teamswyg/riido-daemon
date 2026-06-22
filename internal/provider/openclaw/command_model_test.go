package openclaw

import (
	"slices"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestBuildStartDoesNotEmitUnsupportedModelFlag(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		Model:  "llama3.2:latest",
		Prompt: "hello",
	}, StartOptions{SessionID: "sess-model"})
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(cmd.Args, "--model") {
		t.Fatalf("OpenClaw CLI does not accept --model for agent runs: %v", cmd.Args)
	}
}
