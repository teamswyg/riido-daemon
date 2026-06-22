package mwsdbridge

import (
	"context"
	"fmt"
	"net"
)

func applyRoundTripDeadline(ctx context.Context, conn net.Conn) error {
	deadline, ok := ctx.Deadline()
	if !ok {
		return nil
	}
	if err := conn.SetDeadline(deadline); err != nil {
		return fmt.Errorf("set mwsd socket deadline: %w", err)
	}
	return nil
}
