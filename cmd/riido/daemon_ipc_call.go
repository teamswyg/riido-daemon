package main

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"time"
)

func daemonCall(sock string, method daemonMethod) error {
	conn, err := net.DialTimeout("unix", sock, 2*time.Second)
	if err != nil {
		return daemonWrapf(ErrDaemonSocket, "ipc.call.dial", err, "dial %s", sock)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
	if err := json.NewEncoder(conn).Encode(daemonRequest{Method: method}); err != nil {
		return daemonWrapf(ErrDaemonSocket, "ipc.call.encode", err, "encode request")
	}
	body, err := io.ReadAll(conn)
	if err != nil && !errors.Is(err, io.EOF) {
		return daemonWrapf(ErrDaemonSocket, "ipc.call.read", err, "read response")
	}
	_, err = os.Stdout.Write(body)
	return err
}
