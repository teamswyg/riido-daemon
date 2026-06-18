package mwsdbridge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
)

// Request sends one mwsd method request and decodes the response data.
func (c Client) Request(ctx context.Context, method string, out any) error {
	if c.SocketPath == "" {
		return errors.New("mwsd socket path is empty")
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout())
	defer cancel()

	responseBody, err := c.roundTrip(ctx, Method(method))
	if err != nil {
		return err
	}
	return decodeResponse(method, responseBody, out)
}

func (c Client) roundTrip(ctx context.Context, method Method) ([]byte, error) {
	conn, err := (&net.Dialer{}).DialContext(ctx, "unix", c.SocketPath)
	if err != nil {
		return nil, fmt.Errorf("connect mwsd socket: %w", err)
	}
	defer conn.Close()

	body, err := json.Marshal(request{Method: method})
	if err != nil {
		return nil, fmt.Errorf("encode mwsd request: %w", err)
	}
	if _, err := conn.Write(body); err != nil {
		return nil, fmt.Errorf("write mwsd request: %w", err)
	}
	if unix, ok := conn.(*net.UnixConn); ok {
		if err := unix.CloseWrite(); err != nil {
			return nil, fmt.Errorf("close mwsd request stream: %w", err)
		}
	}

	responseBody, err := io.ReadAll(conn)
	if err != nil {
		return nil, fmt.Errorf("read mwsd response: %w", err)
	}
	return responseBody, nil
}
