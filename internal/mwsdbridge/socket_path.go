package mwsdbridge

import (
	"os"
	"path/filepath"
)

// DefaultSocketPath returns the launchd-backed mwsd socket path.
func DefaultSocketPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Application Support", "macmini-workspace", "mwsd.sock"), nil
}
