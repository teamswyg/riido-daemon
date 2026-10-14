package codex

import (
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestBuildStartBlocksConfigOverrideArgs(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		Cwd: "/tmp/work",
		CustomArgs: []string{
			"-c", "default_permissions=\"unsafe\"",
			"--config=permissions.unsafe.filesystem={\":minimal\"=\"read\"}",
			"--enable", "experimental_untrusted",
			"--disable=some_guard",
			"--safe",
		},
	}, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range configOverrideTokens() {
		if !slices.Contains(cmd.DroppedArgs, want) {
			t.Fatalf("config override arg %q must be dropped: %v", want, cmd.DroppedArgs)
		}
	}
	if containsAnyCommandArg(cmd.Args, "unsafe", "experimental_untrusted", "some_guard") {
		t.Fatalf("config override value bled through: %q", strings.Join(cmd.Args, " "))
	}
	if !slices.Contains(cmd.Args, "--safe") {
		t.Fatalf("non-critical custom arg lost: %v", cmd.Args)
	}
}

func configOverrideTokens() []string {
	return []string{
		"-c",
		"default_permissions=\"unsafe\"",
		"--config=permissions.unsafe.filesystem={\":minimal\"=\"read\"}",
		"--enable",
		"experimental_untrusted",
		"--disable=some_guard",
	}
}
