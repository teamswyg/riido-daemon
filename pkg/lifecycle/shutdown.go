package lifecycle

import (
	"context"
	"strings"
	"time"
)

const (
	DefaultGracefulShutdownTimeout = 5 * time.Second
	DefaultForcedShutdownTimeout   = time.Second
)

func NormalizeShutdownLevel(level ShutdownLevel) ShutdownLevel {
	if level.IsShutdown() {
		return level
	}
	return ShutdownGraceful
}

func ParseShutdownLevel(value string) (ShutdownLevel, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case ShutdownNone.String():
		return ShutdownNone, true
	case ShutdownGraceful.String():
		return ShutdownGraceful, true
	case ShutdownForced.String():
		return ShutdownForced, true
	default:
		return ShutdownNone, false
	}
}

func DefaultShutdownTimeout(level ShutdownLevel) time.Duration {
	if NormalizeShutdownLevel(level).IsForced() {
		return DefaultForcedShutdownTimeout
	}
	return DefaultGracefulShutdownTimeout
}

func StopContext(ctx context.Context) Context {
	lctx := FromContext(ctx)
	level := NormalizeShutdownLevel(lctx.ShutdownLevel())
	if level != lctx.ShutdownLevel() {
		return lctx.WithShutdownLevel(level)
	}
	return lctx
}

func DetachedDefaultShutdown(level ShutdownLevel) (Context, context.CancelFunc) {
	level = NormalizeShutdownLevel(level)
	return DetachedShutdown(level, DefaultShutdownTimeout(level))
}
