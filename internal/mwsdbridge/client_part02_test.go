package mwsdbridge

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestRequestRejectsNotOK(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "mwsd.sock")
	stop := serveRawMwsd(t, socketPath, func(method string) string {
		return `{"ok":false,"method":"` + method + `","data":null,"error":"not ready"}`
	})
	defer stop()

	var status Status
	err := NewClient(socketPath).Request(context.Background(), "status", &status)
	if err == nil {
		t.Fatal("Request should fail when mwsd returns ok=false")
	}
}

func serveFakeMwsd(t *testing.T, socketPath string, data map[string]string) func() {
	t.Helper()
	return serveRawMwsd(t, socketPath, func(method string) string {
		payload, ok := data[method]
		if !ok {
			return `{"ok":false,"method":"` + method + `","data":null,"error":"unknown method"}`
		}
		return `{"ok":true,"method":"` + method + `","data":` + payload + `,"error":null}`
	})
}

func serveRawMwsd(t *testing.T, socketPath string, respond func(method string) string) func() {
	t.Helper()
	_ = os.Remove(socketPath)
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen unix socket: %v", err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				defer conn.Close()
				var req request
				if err := json.NewDecoder(conn).Decode(&req); err != nil {
					return
				}
				_, _ = conn.Write([]byte(respond(req.Method)))
			}(conn)
		}
	}()
	return func() {
		_ = listener.Close()
		<-done
		_ = os.Remove(socketPath)
	}
}
