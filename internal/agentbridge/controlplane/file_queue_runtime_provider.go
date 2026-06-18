package controlplane

import (
	"errors"
	"io/fs"
	"os"
	"strings"
)

func (s *FileQueueSource) runtimeProviderAvailable(runtimeID, provider string) (bool, bool, error) {
	provider = strings.TrimSpace(provider)
	if runtimeID == "" || provider == "" {
		return true, false, nil
	}
	body, err := os.ReadFile(s.runtimePath(runtimeID))
	if errors.Is(err, fs.ErrNotExist) {
		return true, false, nil
	}
	if err != nil {
		return false, false, controlPlaneWrapf(ErrControlPlaneRegistry, "file-queue.runtime-provider-available", err, "read runtime registry")
	}
	rec, err := parseRuntimeRecord(body)
	if err != nil {
		return false, false, err
	}
	key := "provider." + provider + ".available"
	if available, ok := rec.Capabilities[key]; ok {
		return available, true, nil
	}
	if runtimeHasProviderAvailability(rec.Capabilities) {
		return false, true, nil
	}
	return true, false, nil
}

func runtimeHasProviderAvailability(capabilities map[string]bool) bool {
	for capabilityKey := range capabilities {
		if strings.HasPrefix(capabilityKey, "provider.") && strings.HasSuffix(capabilityKey, ".available") {
			return true
		}
	}
	return false
}
