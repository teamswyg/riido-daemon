package riidoapi

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDefaultSocketPathUsesUserHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	got, err := DefaultSocketPath()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, "Library", "Application Support", "riido", "riido.sock")
	if got != want {
		t.Fatalf("DefaultSocketPath() = %q, want %q", got, want)
	}
}

func TestClientTimeoutDefaultAndOverride(t *testing.T) {
	if got := clientTimeout(0); got != 3*time.Second {
		t.Fatalf("default timeout = %s", got)
	}
	if got := clientTimeout(time.Minute); got != time.Minute {
		t.Fatalf("override timeout = %s", got)
	}
}

func TestResponseFailureWithEmptyMessage(t *testing.T) {
	err := responseFailure("status", "")
	if err == nil || err.Error() != "riido API status failed" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateLocalTransportPathErrors(t *testing.T) {
	err := validateLocalTransportPath(LocalTransportUnixSocket, " ")
	if err == nil || !strings.Contains(err.Error(), "endpoint path is empty") {
		t.Fatalf("empty path error = %v", err)
	}
	err = errorsForTransport(LocalTransportWindowsNamedPipe, "boom")
	if err == nil || err.Error() != "windows-named-pipe boom" {
		t.Fatalf("transport error = %v", err)
	}
}
