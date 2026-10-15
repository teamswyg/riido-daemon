//go:build windows

package riidoapi

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"
)

func TestWindowsNamedPipeClosePreservesBufferedResponse(t *testing.T) {
	pipe := fmt.Sprintf(`\\.\pipe\riido-delayed-read-%d`, time.Now().UnixNano())
	listener := newNamedPipeListener(pipe)
	defer listener.Close()
	payload := strings.Repeat("buffered-response-", 32)
	serverDone := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			serverDone <- err
			return
		}
		_, err = io.WriteString(conn, payload)
		closeErr := conn.Close()
		if err == nil {
			err = closeErr
		}
		serverDone <- err
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	conn, err := dialNamedPipe(ctx, pipe)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	// Deliberately read only after the server has closed. Explicit disconnect
	// discarded the entire buffered response, and a flush here would deadlock.
	select {
	case err := <-serverDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("server close blocked on a client that has not read")
	}
	body, err := io.ReadAll(conn)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != payload {
		t.Fatalf("delayed client received %d of %d bytes", len(body), len(payload))
	}
}
