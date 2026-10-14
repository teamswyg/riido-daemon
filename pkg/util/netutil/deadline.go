package netutil

import (
	"context"
	"fmt"
	"net"
)

func ApplyContextDeadline(ctx context.Context, conn net.Conn, surface string) error {
	deadline, ok := ctx.Deadline()
	if !ok {
		return nil
	}
	if err := conn.SetDeadline(deadline); err != nil {
		return fmt.Errorf("set %s deadline: %w", surface, err)
	}
	return nil
}
