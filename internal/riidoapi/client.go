package riidoapi

import (
	"os"
	"path/filepath"
	"time"
)

type Client struct {
	SocketPath string
	Transport  LocalTransport
	Timeout    time.Duration
}

func DefaultSocketPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Application Support", "riido", "riido.sock"), nil
}

func NewClient(socketPath string) Client {
	return NewClientWithTransport(LocalTransportUnixSocket, socketPath)
}

func NewClientWithTransport(transport LocalTransport, socketPath string) Client {
	return Client{
		SocketPath: socketPath,
		Transport:  normalizeLocalTransport(transport),
		Timeout:    3 * time.Second,
	}
}
