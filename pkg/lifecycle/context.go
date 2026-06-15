package lifecycle

import (
	"context"
	"os"
	"os/signal"
	"time"
)

type ShutdownLevel uint8

const (
	ShutdownNone ShutdownLevel = iota
	ShutdownGraceful
	ShutdownForced
)

func (l ShutdownLevel) String() string {
	switch l {
	case ShutdownNone:
		return "none"
	case ShutdownGraceful:
		return "graceful"
	case ShutdownForced:
		return "forced"
	default:
		return "unknown"
	}
}

func (l ShutdownLevel) AtLeast(want ShutdownLevel) bool {
	return l >= want
}

func (l ShutdownLevel) IsShutdown() bool {
	return l.AtLeast(ShutdownGraceful)
}

func (l ShutdownLevel) IsForced() bool {
	return l.AtLeast(ShutdownForced)
}

type shutdownLevelKey struct{}

// Context carries daemon lifecycle semantics separately from stdlib
// context.Context. It intentionally does not implement context.Context; pass
// Context() explicitly when crossing into stdlib or external library APIs.
type Context struct {
	ctx   context.Context
	level ShutdownLevel
}

func Background() Context {
	return New(context.Background(), ShutdownNone)
}

func FromContext(ctx context.Context) Context {
	if ctx == nil {
		return Background()
	}
	if level, ok := ctx.Value(shutdownLevelKey{}).(ShutdownLevel); ok {
		return Context{ctx: ctx, level: level}
	}
	return New(ctx, ShutdownNone)
}

func New(ctx context.Context, level ShutdownLevel) Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return Context{
		ctx:   context.WithValue(ctx, shutdownLevelKey{}, level),
		level: level,
	}
}

func (c Context) Context() context.Context {
	if c.ctx == nil {
		return Background().ctx
	}
	return c.ctx
}

func (c Context) ShutdownLevel() ShutdownLevel {
	return c.level
}

func (c Context) WithShutdownLevel(level ShutdownLevel) Context {
	return New(c.Context(), level)
}

func (c Context) Done() <-chan struct{} {
	return c.Context().Done()
}

func (c Context) Err() error {
	return c.Context().Err()
}

func WithCancel(parent Context) (Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent.Context())
	return New(ctx, parent.ShutdownLevel()), cancel
}

func WithTimeout(parent Context, timeout time.Duration) (Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(parent.Context(), timeout)
	return New(ctx, parent.ShutdownLevel()), cancel
}

func Notify(parent Context, signals ...os.Signal) (Context, func()) {
	ctx, stop := signal.NotifyContext(parent.Context(), signals...)
	return New(ctx, parent.ShutdownLevel()), stop
}

func DetachedShutdown(level ShutdownLevel, timeout time.Duration) (Context, context.CancelFunc) {
	return WithTimeout(New(context.Background(), level), timeout)
}
