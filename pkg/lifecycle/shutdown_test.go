package lifecycle

import (
	"context"
	"testing"
)

func TestNormalizeShutdownLevelDefaultsToGraceful(t *testing.T) {
	if got := NormalizeShutdownLevel(ShutdownNone); got != ShutdownGraceful {
		t.Fatalf("NormalizeShutdownLevel(none) = %s, want %s", got, ShutdownGraceful)
	}
	if got := NormalizeShutdownLevel(ShutdownForced); got != ShutdownForced {
		t.Fatalf("NormalizeShutdownLevel(forced) = %s, want %s", got, ShutdownForced)
	}
}

func TestParseShutdownLevel(t *testing.T) {
	for _, tc := range []struct {
		raw   string
		level ShutdownLevel
		ok    bool
	}{
		{raw: "none", level: ShutdownNone, ok: true},
		{raw: " graceful ", level: ShutdownGraceful, ok: true},
		{raw: "FORCED", level: ShutdownForced, ok: true},
		{raw: "", ok: false},
		{raw: "bogus", ok: false},
	} {
		got, ok := ParseShutdownLevel(tc.raw)
		if ok != tc.ok || got != tc.level {
			t.Fatalf("ParseShutdownLevel(%q) = (%s, %v), want (%s, %v)", tc.raw, got, ok, tc.level, tc.ok)
		}
	}
}

func TestDefaultShutdownTimeoutByLevel(t *testing.T) {
	if got := DefaultShutdownTimeout(ShutdownGraceful); got != DefaultGracefulShutdownTimeout {
		t.Fatalf("graceful timeout = %s, want %s", got, DefaultGracefulShutdownTimeout)
	}
	if got := DefaultShutdownTimeout(ShutdownForced); got != DefaultForcedShutdownTimeout {
		t.Fatalf("forced timeout = %s, want %s", got, DefaultForcedShutdownTimeout)
	}
}

func TestStopContextNormalizesLevel(t *testing.T) {
	got := StopContext(context.Background())
	if got.ShutdownLevel() != ShutdownGraceful {
		t.Fatalf("StopContext level = %s, want %s", got.ShutdownLevel(), ShutdownGraceful)
	}

	forced := StopContext(New(context.Background(), ShutdownForced).Context())
	if forced.ShutdownLevel() != ShutdownForced {
		t.Fatalf("forced StopContext level = %s, want %s", forced.ShutdownLevel(), ShutdownForced)
	}
}
