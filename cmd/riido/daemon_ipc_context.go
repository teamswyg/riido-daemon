package main

import (
	"context"
	"net"
)

func daemonConnClosedContext(conn net.Conn) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		var one [1]byte
		for {
			if _, err := conn.Read(one[:]); err != nil {
				cancel()
				return
			}
		}
	}()
	return ctx, cancel
}
