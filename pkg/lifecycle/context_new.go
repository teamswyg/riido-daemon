package lifecycle

import "context"

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
