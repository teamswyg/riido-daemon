package openclaw

import (
	"context"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestDetectMissingBinary(t *testing.T) {
	res, err := Detect(context.Background(), agentbridge.DetectEnv{
		EnvOverride: map[string]string{EnvOverride: "/no/such/openclaw"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Available {
		t.Fatalf("Available: %+v", res)
	}
}

func TestDetectAcceptsAtMinimumVersion(t *testing.T) {
	exe := writeShim(t, MinSupportedVersion)
	res, _ := Detect(context.Background(), agentbridge.DetectEnv{
		EnvOverride: map[string]string{EnvOverride: exe},
	})
	if !res.Available {
		t.Fatalf("Available: %+v", res)
	}
}

func TestDetectRejectsOlderThanMinimum(t *testing.T) {
	exe := writeShim(t, "2026.4.30")
	res, _ := Detect(context.Background(), agentbridge.DetectEnv{
		EnvOverride: map[string]string{EnvOverride: exe},
	})
	if res.Available {
		t.Fatalf("expected gate to reject older version: %+v", res)
	}
	if !strings.Contains(res.Reason, MinSupportedVersion) {
		t.Fatalf("reason should mention minimum: %q", res.Reason)
	}
}

func TestDetectAcceptsNewerVersion(t *testing.T) {
	exe := writeShim(t, "v2026.12.31")
	res, _ := Detect(context.Background(), agentbridge.DetectEnv{
		EnvOverride: map[string]string{EnvOverride: exe},
	})
	if !res.Available {
		t.Fatalf("newer should pass: %+v", res)
	}
}

func TestDetectUnparseableVersion(t *testing.T) {
	exe := writeShim(t, "garbage-version")
	res, _ := Detect(context.Background(), agentbridge.DetectEnv{
		EnvOverride: map[string]string{EnvOverride: exe},
	})
	if res.Available {
		t.Fatalf("unparseable version must not be Available: %+v", res)
	}
}
