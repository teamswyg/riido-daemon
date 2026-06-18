package openclaw

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestBuildStartRequiresSessionID(t *testing.T) {
	_, err := BuildStart(agentbridge.StartRequest{}, StartOptions{})
	if err == nil {
		t.Fatal("expected error without session id")
	}
}

func TestBuildStartDerivesSessionIDFromTaskID(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		TaskID: "task-openclaw-1",
		Prompt: "do the thing",
	}, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if args := strings.Join(cmd.Args, " "); !strings.Contains(args, "--session-id task-openclaw-1") {
		t.Fatalf("session id not derived from task id: %q", args)
	}
}

func TestBuildStartPrefersResumeSessionID(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		TaskID:          "task-openclaw-1",
		ResumeSessionID: "sess-existing",
		Prompt:          "continue",
	}, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if args := strings.Join(cmd.Args, " "); !strings.Contains(args, "--session-id sess-existing") {
		t.Fatalf("resume session id not preferred: %q", args)
	}
}
