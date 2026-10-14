package lifecycle

import (
	"context"
	"os"
	"os/signal"
	"time"
)

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
