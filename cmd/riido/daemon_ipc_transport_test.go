package main

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/riidoapi"
)

func TestDaemonLocalTransportForOS(t *testing.T) {
	tests := []struct {
		goos string
		want riidoapi.LocalTransport
	}{
		{goos: "windows", want: riidoapi.LocalTransportWindowsNamedPipe},
		{goos: "darwin", want: riidoapi.LocalTransportUnixSocket},
		{goos: "linux", want: riidoapi.LocalTransportUnixSocket},
	}
	for _, test := range tests {
		if got := daemonLocalTransportForOS(test.goos); got != test.want {
			t.Errorf("daemonLocalTransportForOS(%q) = %q, want %q", test.goos, got, test.want)
		}
	}
}
