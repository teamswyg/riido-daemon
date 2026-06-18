package detectutil

import (
	"context"
	"testing"
)

func TestVersionProbeEchoesOutput(t *testing.T) {
	out, ok := VersionProbe(context.Background(), "/bin/echo", "1.2.3")
	if !ok {
		t.Fatal("probe failed")
	}
	if out != "1.2.3" {
		t.Fatalf("output: %q", out)
	}
}

func TestVersionProbeMissingBinary(t *testing.T) {
	_, ok := VersionProbe(context.Background(), "", "--version")
	if ok {
		t.Fatal("empty exe should fail")
	}
	_, ok = VersionProbe(context.Background(), "/no/such/path", "--version")
	if ok {
		t.Fatal("missing binary should fail")
	}
}
