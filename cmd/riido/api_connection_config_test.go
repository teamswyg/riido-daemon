package main

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/riidoapi"
)

func TestDefaultAPICLIConfigForWindowsNamedPipe(t *testing.T) {
	userHome := func() (string, error) { return `C:\Users\tester`, nil }
	localAppData := func() (string, error) { return `C:\Users\tester\AppData\Local`, nil }

	config, err := defaultAPICLIConfigForOS("windows", userHome, localAppData)
	if err != nil {
		t.Fatal(err)
	}
	if config.transport != riidoapi.LocalTransportWindowsNamedPipe {
		t.Fatalf("transport = %q", config.transport)
	}
	if config.socketPath != `\\.\pipe\riido-dev-local-helper-riido` {
		t.Fatalf("socket path = %q", config.socketPath)
	}
}

func TestDefaultAPICLIConfigPreservesDarwinSocket(t *testing.T) {
	userHome := func() (string, error) { return "/Users/tester", nil }
	unusedCache := func() (string, error) {
		t.Fatal("Darwin defaults must not resolve Windows local app data")
		return "", nil
	}

	config, err := defaultAPICLIConfigForOS("darwin", userHome, unusedCache)
	if err != nil {
		t.Fatal(err)
	}
	if config.transport != riidoapi.LocalTransportUnixSocket {
		t.Fatalf("transport = %q", config.transport)
	}
	if config.socketPath != "/Users/tester/Library/Application Support/riido/riido.sock" {
		t.Fatalf("socket path = %q", config.socketPath)
	}
}
