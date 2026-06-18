package claude

import (
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestBuildStartDropsBlockedCustomArgs(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		CustomArgs: []string{
			"--output-format", "text",
			"--permission-mode=bypassPermissions",
			"--my-flag", "ok",
		},
	}, safeStartOptions())
	if err != nil {
		t.Fatal(err)
	}
	args := strings.Join(cmd.Args, " ")
	if strings.HasSuffix(args, "--output-format text") || strings.Contains(args, "--output-format text ") {
		t.Fatalf("caller overrode --output-format: %q", args)
	}
	if strings.Contains(args, "--permission-mode=bypassPermissions") {
		t.Fatalf("caller's bypassPermissions bled through: %q", args)
	}
	if !strings.Contains(args, "--my-flag ok") {
		t.Fatalf("non-blocked custom arg lost: %q", args)
	}
	assertDroppedArgs(t, cmd.DroppedArgs, "--output-format", "--permission-mode=bypassPermissions")
}

func assertDroppedArgs(t *testing.T, got []string, wants ...string) {
	t.Helper()
	if len(got) == 0 {
		t.Fatal("expected dropped args to be surfaced, got none")
	}
	for _, want := range wants {
		if !slices.Contains(got, want) {
			t.Fatalf("expected %q in DroppedArgs %v", want, got)
		}
	}
}
