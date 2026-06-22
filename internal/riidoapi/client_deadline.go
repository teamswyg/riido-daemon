package riidoapi

import (
	"context"
	"fmt"
	"net"
)

func applyClientDeadline(ctx context.Context, conn net.Conn) error {
	deadline, ok := ctx.Deadline()
	if !ok {
		return nil
	}
	if err := conn.SetDeadline(deadline); err != nil {
		return fmt.Errorf("set riido API endpoint deadline: %w", err)
	}
	return nil
}
