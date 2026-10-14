//go:build windows

package riidoapi

import (
	"errors"
	"strings"
)

func validateWindowsNamedPipePath(path string) error {
	if !strings.HasPrefix(strings.ToLower(path), `\\.\pipe\`) {
		return errors.New(`windows named pipe path must start with \\.\pipe\`)
	}
	return nil
}
