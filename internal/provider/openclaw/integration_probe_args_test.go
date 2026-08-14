package openclaw

import (
	"strconv"
	"testing"
)

func configProbeArgs(runNonce int64) []string {
	args := []string{
		"agent",
		"--local",
		"--json",
		"--session-id",
		"riido-config-probe-" + strconv.FormatInt(runNonce, 10),
		"--message",
		"Say OK only.",
		"--timeout",
		"30",
	}
	return args
}

func TestConfigProbeArgsUsesFreshSessionNonce(t *testing.T) {
	t.Parallel()

	first := configProbeArgs(101)
	second := configProbeArgs(102)
	if got, want := first[4], "riido-config-probe-101"; got != want {
		t.Fatalf("first session id = %q, want %q", got, want)
	}
	if first[4] == second[4] {
		t.Fatalf("config probe reused session id %q", first[4])
	}
}
