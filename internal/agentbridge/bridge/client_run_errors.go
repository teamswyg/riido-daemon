package bridge

import "fmt"

func providerUnavailableError(provider Provider, reason string) error {
	return fmt.Errorf("bridge: provider %s unavailable: %s", provider, reason)
}
