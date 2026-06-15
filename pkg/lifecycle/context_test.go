package lifecycle

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestShutdownLevelOrdering(t *testing.T) {
	if !ShutdownForced.AtLeast(ShutdownGraceful) {
		t.Fatal("forced shutdown should satisfy graceful threshold")
	}
	if ShutdownGraceful.IsForced() {
		t.Fatal("graceful shutdown should not be forced")
	}
}

func TestContextRoundTripThroughStdlibContext(t *testing.T) {
	lctx := New(context.Background(), ShutdownGraceful)

	got := FromContext(lctx.Context())
	if got.ShutdownLevel() != ShutdownGraceful {
		t.Fatalf("ShutdownLevel() = %s, want %s", got.ShutdownLevel(), ShutdownGraceful)
	}
}

func TestContextDoesNotImplementStdlibContext(t *testing.T) {
	stdlibContext := reflect.TypeFor[context.Context]()
	if reflect.TypeFor[Context]().Implements(stdlibContext) {
		t.Fatal("lifecycle.Context must require explicit Context() conversion")
	}
}

func TestDetachedShutdownTimesOut(t *testing.T) {
	lctx, cancel := DetachedShutdown(ShutdownForced, time.Nanosecond)
	defer cancel()

	select {
	case <-lctx.Done():
	case <-time.After(time.Second):
		t.Fatal("detached shutdown context did not time out")
	}
	if lctx.ShutdownLevel() != ShutdownForced {
		t.Fatalf("ShutdownLevel() = %s, want %s", lctx.ShutdownLevel(), ShutdownForced)
	}
}
