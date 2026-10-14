package detectutil

import (
	"context"
	"testing"
)

func TestVersionProbeStrictReportsExitCodeAndOutput(t *testing.T) {
	res := VersionProbeStrict(context.Background(), "/bin/sh", "-c", "printf 'tool failed'; exit 7")
	if !res.OK {
		t.Fatal("strict probe should report command completion")
	}
	if res.ExitCode != 7 {
		t.Fatalf("exit code: %d", res.ExitCode)
	}
	if res.Output != "tool failed" {
		t.Fatalf("output: %q", res.Output)
	}
}

func TestVersionProbeStrictMissingBinary(t *testing.T) {
	res := VersionProbeStrict(context.Background(), "/no/such/path", "--version")
	if res.OK {
		t.Fatalf("missing binary should fail closed: %+v", res)
	}
}
