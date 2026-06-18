package openclaw

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestBuildStartSystemPromptInlineFallback(t *testing.T) {
	cmd, _ := BuildStart(agentbridge.StartRequest{
		Prompt:       "user task",
		SystemPrompt: "be careful",
	}, StartOptions{SessionID: "sess-2"})
	msgFlag := false
	for i, a := range cmd.Args {
		if a == "--message" && i+1 < len(cmd.Args) {
			if !strings.Contains(cmd.Args[i+1], "be careful") || !strings.Contains(cmd.Args[i+1], "user task") {
				t.Fatalf("inline fallback missing content: %q", cmd.Args[i+1])
			}
			msgFlag = true
		}
	}
	if !msgFlag {
		t.Fatalf("--message not built: %v", cmd.Args)
	}
}
