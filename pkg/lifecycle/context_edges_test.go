package lifecycle

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestBackgroundAndNilContextDefaults(t *testing.T) {
	if got := Background(); got.ShutdownLevel() != ShutdownNone || got.Err() != nil {
		t.Fatalf("unexpected background lifecycle context: level=%s err=%v", got.ShutdownLevel(), got.Err())
	}
	if got := FromContext(nil); got.ShutdownLevel() != ShutdownNone {
		t.Fatalf("nil stdlib context level = %s", got.ShutdownLevel())
	}
	if got := (Context{}).Context().Err(); got != nil {
		t.Fatalf("zero lifecycle context should use live background, got %v", got)
	}
}

func TestWithCancelPreservesLevelAndCancels(t *testing.T) {
	lctx, cancel := WithCancel(New(context.Background(), ShutdownForced))
	if lctx.ShutdownLevel() != ShutdownForced {
		t.Fatalf("level = %s, want forced", lctx.ShutdownLevel())
	}
	cancel()
	select {
	case <-lctx.Done():
	case <-time.After(time.Second):
		t.Fatal("cancel did not close lifecycle context")
	}
	if lctx.Err() != context.Canceled {
		t.Fatalf("Err() = %v, want context.Canceled", lctx.Err())
	}
}

func TestNotifyStopCancelsAndPreservesLevel(t *testing.T) {
	lctx, stop := Notify(New(context.Background(), ShutdownGraceful), os.Interrupt)
	if lctx.ShutdownLevel() != ShutdownGraceful {
		t.Fatalf("level = %s, want graceful", lctx.ShutdownLevel())
	}
	stop()
	select {
	case <-lctx.Done():
	case <-time.After(time.Second):
		t.Fatal("Notify stop did not close context")
	}
}

func TestDetachedDefaultShutdownUsesNormalizedTimeout(t *testing.T) {
	for _, tc := range []struct {
		level ShutdownLevel
		want  time.Duration
	}{
		{level: ShutdownNone, want: DefaultGracefulShutdownTimeout},
		{level: ShutdownForced, want: DefaultForcedShutdownTimeout},
	} {
		lctx, cancel := DetachedDefaultShutdown(tc.level)
		defer cancel()
		if DefaultShutdownTimeout(lctx.ShutdownLevel()) != tc.want {
			t.Fatalf("timeout for %s did not normalize to %s", tc.level, tc.want)
		}
	}
}

func TestUnknownShutdownLevelString(t *testing.T) {
	if got := ShutdownLevel(99).String(); got != "unknown" {
		t.Fatalf("unknown level string = %q", got)
	}
}
