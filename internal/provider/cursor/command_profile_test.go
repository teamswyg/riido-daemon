package cursor

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestBuildStartProfileAgentSubcommand(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{Cwd: "/tmp/work", Prompt: "hi"}, StartOptions{Profile: ProfileAgentSubcommand})
	if err != nil {
		t.Fatal(err)
	}
	if len(cmd.Args) == 0 || cmd.Args[0] != "agent" {
		t.Fatalf("agent-subcommand profile must start with 'agent', got %v", cmd.Args)
	}
}

func TestBuildStartProfileLegacyChatOptIn(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{Cwd: "/tmp/work", Prompt: "hi"}, StartOptions{Profile: ProfileLegacyChat})
	if err != nil {
		t.Fatal(err)
	}
	if len(cmd.Args) == 0 || cmd.Args[0] != "chat" {
		t.Fatalf("legacy-chat profile must start with 'chat', got %v", cmd.Args)
	}
}

func TestBuildStartProfileUnknownRejected(t *testing.T) {
	_, err := BuildStart(agentbridge.StartRequest{Prompt: "x"}, StartOptions{Profile: "ghost"})
	if err == nil {
		t.Fatal("expected error for unknown profile")
	}
}
