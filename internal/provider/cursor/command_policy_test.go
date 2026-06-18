package cursor

import (
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestBuildStartBlockedArgs(t *testing.T) {
	for _, want := range []string{"-p", "--output-format", "--yolo"} {
		if !slices.Contains(BlockedArgs(), want) {
			t.Fatalf("BlockedArgs missing %q: %v", want, BlockedArgs())
		}
	}
}

func TestBuildStartUnsupportedSystemPromptSurfaceWarning(t *testing.T) {
	cmd, _ := BuildStart(agentbridge.StartRequest{
		Prompt:       "x",
		SystemPrompt: "be careful",
		MaxTurns:     5,
	}, StartOptions{})
	if len(cmd.DroppedArgs) == 0 {
		t.Fatalf("expected DroppedArgs to record unsupported features, got none")
	}
	joined := strings.Join(cmd.DroppedArgs, " ")
	if !strings.Contains(joined, "system_prompt") {
		t.Fatalf("system_prompt not surfaced: %v", cmd.DroppedArgs)
	}
	if !strings.Contains(joined, "max_turns") {
		t.Fatalf("max_turns not surfaced: %v", cmd.DroppedArgs)
	}
}
