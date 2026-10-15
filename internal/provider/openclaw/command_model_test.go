package openclaw

import (
	"slices"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestBuildStartPinsModelWithAgentExecFlag(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		Model:  "llama3.2:latest",
		Prompt: "hello",
	}, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	index := slices.Index(cmd.Args, "--model")
	if index < 0 || index+1 >= len(cmd.Args) || cmd.Args[index+1] != "llama3.2:latest" {
		t.Fatalf("OpenClaw agent exec ignored selected model: %v", cmd.Args)
	}
}
