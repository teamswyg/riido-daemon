package riidoapi

import (
	"context"
	"errors"
	"fmt"
)

func (s Server) Serve(ctx context.Context) error {
	transport := normalizeLocalTransport(s.config.Transport)
	if s.config.SocketPath == "" {
		return errors.New("riido API socket path is empty")
	}
	if s.config.TaskDBPath == "" {
		return errors.New("riido task DB path is empty")
	}
	listener, cleanup, err := listenLocalEndpoint(transport, s.config.SocketPath)
	if err != nil {
		return fmt.Errorf("listen riido API %s endpoint: %w", transport, err)
	}
	defer cleanup()

	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("accept riido API connection: %w", err)
		}
		go s.handleConn(ctx, conn)
	}
}
