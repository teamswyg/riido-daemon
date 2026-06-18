package claude

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestBuildStartResumeMaxTurnsModelSystem(t *testing.T) {
	cmd, _ := BuildStart(agentbridge.StartRequest{
		Model:           "claude-opus-4-7",
		SystemPrompt:    "be concise",
		MaxTurns:        4,
		ResumeSessionID: "sess-abc",
	}, safeStartOptions())
	args := strings.Join(cmd.Args, " ")
	for _, want := range []string{
		"--model claude-opus-4-7",
		"--append-system-prompt be concise",
		"--max-turns 4",
		"--resume sess-abc",
	} {
		if !strings.Contains(args, want) {
			t.Fatalf("missing arg %q in %q", want, args)
		}
	}
}
