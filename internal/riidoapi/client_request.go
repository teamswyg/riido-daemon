package riidoapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/teamswyg/riido-daemon/pkg/util/netutil"
)

func (c Client) Request(ctx context.Context, method string, params, out any) error {
	transport := normalizeLocalTransport(c.Transport)
	if c.SocketPath == "" {
		return errors.New("riido API socket path is empty")
	}
	ctx, cancel := context.WithTimeout(ctx, clientTimeout(c.Timeout))
	defer cancel()

	conn, err := dialLocalEndpoint(ctx, transport, c.SocketPath)
	if err != nil {
		return fmt.Errorf("connect riido API %s endpoint: %w", transport, err)
	}
	defer conn.Close()
	if err := netutil.ApplyContextDeadline(ctx, conn, "riido API endpoint"); err != nil {
		return err
	}
	if err := writeRequest(conn, method, params); err != nil {
		return err
	}
	responseBody, err := io.ReadAll(conn)
	if err != nil {
		return fmt.Errorf("read riido API response: %w", err)
	}
	return decodeResponse(responseBody, method, out)
}

func writeRequest(conn net.Conn, method string, params any) error {
	rawParams, err := rawParams(params)
	if err != nil {
		return err
	}
	requestBody, err := json.Marshal(requestEnvelope{Method: Method(method), Params: rawParams})
	if err != nil {
		return fmt.Errorf("encode riido API request: %w", err)
	}
	if _, err := conn.Write(requestBody); err != nil {
		return fmt.Errorf("write riido API request: %w", err)
	}
	if unix, ok := conn.(*net.UnixConn); ok {
		if err := unix.CloseWrite(); err != nil {
			return fmt.Errorf("close riido API request stream: %w", err)
		}
	}
	return nil
}

func clientTimeout(timeout time.Duration) time.Duration {
	if timeout == 0 {
		return 3 * time.Second
	}
	return timeout
}
