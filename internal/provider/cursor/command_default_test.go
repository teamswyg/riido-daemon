package cursor

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestBuildStartDefaultProfileIsRootPrint(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{Cwd: "/tmp/work", Prompt: "do the thing"}, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	args := strings.Join(cmd.Args, " ")
	for _, want := range []string{"-p do the thing", "--output-format stream-json", "--workspace /tmp/work", "--trust"} {
		if !strings.Contains(args, want) {
			t.Fatalf("missing %q in %q", want, args)
		}
	}
	if cmd.Args[0] == "chat" {
		t.Fatalf("default profile must NOT use legacy `chat` subcommand: %v", cmd.Args)
	}
	if cmd.Dir != "/tmp/work" {
		t.Fatalf("Dir: %q", cmd.Dir)
	}
}

func TestBuildStartTrustsDaemonWorkspaceWithoutYolo(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{Cwd: "/tmp/work", Prompt: "do the thing"}, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	args := strings.Join(cmd.Args, " ")
	if !strings.Contains(args, "--trust") {
		t.Fatalf("headless workspace must be trusted explicitly: %v", cmd.Args)
	}
	if strings.Contains(args, "--yolo") {
		t.Fatalf("--trust must not imply unsafe --yolo: %v", cmd.Args)
	}
}

func TestBuildStartUsesRuntimeSelectedExecutable(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		Executable: "/opt/riido/bin/cursor-selected",
		Prompt:     "do the thing",
	}, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Executable != "/opt/riido/bin/cursor-selected" {
		t.Fatalf("runtime-selected executable lost: %q", cmd.Executable)
	}
}
