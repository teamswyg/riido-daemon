package openclaw

import "testing"

func configProbeArgs(_ int64) []string {
	args := []string{
		"agent",
		"exec",
		"--json",
		"--timeout",
		"30",
		"Say OK only.",
	}
	return args
}

func TestConfigProbeArgsUsesIsolatedAgentExec(t *testing.T) {
	t.Parallel()

	got := configProbeArgs(101)
	if len(got) < 3 || got[0] != "agent" || got[1] != "exec" || got[2] != "--json" {
		t.Fatalf("config probe is not isolated agent exec: %v", got)
	}
}
