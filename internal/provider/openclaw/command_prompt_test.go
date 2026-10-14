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
	}, StartOptions{})
	msg := cmd.Args[len(cmd.Args)-1]
	if !strings.Contains(msg, "be careful") || !strings.Contains(msg, "user task") {
		t.Fatalf("inline fallback missing content: %q", msg)
	}
}
