package lifecycle

import "context"

type shutdownLevelKey struct{}

// Context carries daemon lifecycle semantics separately from stdlib
// context.Context. It intentionally does not implement context.Context; pass
// Context() explicitly when crossing into stdlib or external library APIs.
type Context struct {
	ctx   context.Context
	level ShutdownLevel
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
