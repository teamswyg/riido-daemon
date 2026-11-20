package riidoapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"
	"time"
)

func TestAPIResponseSurvivesPeerCloseBeforeRequestHalfClose(t *testing.T) {
	valid := fmt.Sprintf(`{"ok":true,"method":%q,"data":{"app_version":"race.v1"}}`, MethodStatus)
	for _, tc := range []struct {
		name       string
		response   string
		decodeFail bool
		localClose bool
	}{
		{name: "complete response", response: valid},
		{name: "missing response", decodeFail: true},
		{name: "malformed response", response: "{broken", decodeFail: true},
		{name: "mismatched method", response: `{"ok":true,"method":"other","data":{}}`, decodeFail: true},
		{name: "refused response", response: `{"ok":false,"error":"refused"}`, decodeFail: true},
		{name: "invalid data", response: fmt.Sprintf(`{"ok":true,"method":%q,"data":"invalid"}`, MethodStatus), decodeFail: true},
		{name: "local close remains an error", response: valid, localClose: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			conn := connectedAPIResponseBeforeHalfClose(t, tc.response)
			if tc.localClose {
				_ = conn.Close()
				if err := closeRequestStream(conn); !errors.Is(err, net.ErrClosed) {
					t.Fatalf("local close error = %v, want net.ErrClosed", err)
				}
				return
			}
			if err := closeRequestStream(conn); err != nil {
				t.Fatalf("peer close discarded buffered response: %v", err)
			}
			body, err := io.ReadAll(conn)
			if err != nil {
				t.Fatalf("read buffered response: %v", err)
			}
			var status Status
			err = decodeResponse(body, string(MethodStatus), &status)
			if (err != nil) != tc.decodeFail {
				t.Fatalf("decode error = %v, want failure %t", err, tc.decodeFail)
			}
			if !tc.decodeFail && status.AppVersion != "race.v1" {
				t.Fatalf("decoded status = %+v", status)
			}
		})
	}
}

func connectedAPIResponseBeforeHalfClose(t *testing.T, response string) *net.UnixConn {
	t.Helper()
	address := &net.UnixAddr{Name: shortRiidoAPISocketPath(t), Net: "unix"}
	listener, err := net.ListenUnix("unix", address)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	finished := make(chan error, 1)
	go func() {
		server, err := listener.AcceptUnix()
		if err != nil {
			finished <- err
			return
		}
		_ = server.SetDeadline(time.Now().Add(time.Second))
		var request requestEnvelope
		err = json.NewDecoder(server).Decode(&request)
		if err == nil {
			_, err = server.Write([]byte(response))
		}
		_ = server.Close()
		finished <- err
	}()
	conn, err := net.DialUnix("unix", nil, address)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	_ = conn.SetDeadline(time.Now().Add(time.Second))
	if _, err := conn.Write([]byte(`{"method":"status"}`)); err != nil {
		t.Fatal(err)
	}
	if err := <-finished; err != nil {
		t.Fatal(err)
	}
	return conn
}
