package hostintegration

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

func validateEndpointRoot(in LocalIPCEndpointInput) error {
	if strings.TrimSpace(in.AppDataRoot.Path) == "" {
		return errors.New("local IPC endpoint requires app data root")
	}
	if in.AppDataRoot.Channel != in.Channel {
		return fmt.Errorf("app data root channel %q does not match endpoint channel %q", in.AppDataRoot.Channel, in.Channel)
	}
	if in.AppDataRoot.HostOS != in.HostOS {
		return fmt.Errorf("app data root host OS %q does not match endpoint host OS %q", in.AppDataRoot.HostOS, in.HostOS)
	}
	return nil
}

func normalizedEndpointName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		name = "riido"
	}
	if !endpointNameRunesSafe(name) || strings.ContainsAny(name, `/\:`) {
		return "", fmt.Errorf("invalid local IPC endpoint name %q", raw)
	}
	return name, nil
}

func endpointNameRunesSafe(name string) bool {
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == '.' {
			continue
		}
		return false
	}
	return true
}
