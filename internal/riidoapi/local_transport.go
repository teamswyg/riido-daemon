package riidoapi

import (
	"fmt"
	"strings"
)

func normalizeLocalTransport(transport LocalTransport) LocalTransport {
	if strings.TrimSpace(string(transport)) == "" {
		return LocalTransportUnixSocket
	}
	return transport
}

func validateLocalTransportPath(transport LocalTransport, path string) error {
	if strings.TrimSpace(path) == "" {
		return errorsForTransport(transport, "endpoint path is empty")
	}
	switch transport {
	case LocalTransportUnixSocket, LocalTransportWindowsNamedPipe:
		return nil
	default:
		return fmt.Errorf("unknown local transport %q", transport)
	}
}

func errorsForTransport(transport LocalTransport, message string) error {
	return fmt.Errorf("%s %s", transport, message)
}
