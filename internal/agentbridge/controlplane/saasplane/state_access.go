package saasplane

import (
	"context"
	"errors"
)

func (p *Plane) withState(ctx context.Context, fn func(*planeState)) error {
	ack := make(chan struct{})
	select {
	case p.ops <- stateOp{fn: fn, ack: ack}:
	case <-p.done:
		return errors.New("saasplane: closed")
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case <-ack:
		return nil
	case <-p.done:
		return errors.New("saasplane: closed")
	case <-ctx.Done():
		return ctx.Err()
	}
}
