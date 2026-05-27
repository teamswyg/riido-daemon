package riidoapi

import (
	"context"
	"runtime"
	"strings"
	"testing"
)

func TestNormalizeLocalTransportDefaultsToUnixSocket(t *testing.T) {
	if got := normalizeLocalTransport(""); got != LocalTransportUnixSocket {
		t.Fatalf("default transport = %q, want %q", got, LocalTransportUnixSocket)
	}
}

func TestLocalTransportRejectsUnknownTransport(t *testing.T) {
	if _, _, err := listenLocalEndpoint(LocalTransport("tcp"), "127.0.0.1:0"); err == nil {
		t.Fatal("expected unknown transport error")
	}
}

func TestWindowsNamedPipeTransportRequiresWindowsHost(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows named pipe transport is exercised by the Windows build")
	}
	_, _, listenErr := listenLocalEndpoint(LocalTransportWindowsNamedPipe, `\\.\pipe\riido-test`)
	if listenErr == nil || !strings.Contains(listenErr.Error(), "requires Windows") {
		t.Fatalf("listen error = %v, want Windows requirement", listenErr)
	}
	_, dialErr := dialLocalEndpoint(context.Background(), LocalTransportWindowsNamedPipe, `\\.\pipe\riido-test`)
	if dialErr == nil || !strings.Contains(dialErr.Error(), "requires Windows") {
		t.Fatalf("dial error = %v, want Windows requirement", dialErr)
	}
}
