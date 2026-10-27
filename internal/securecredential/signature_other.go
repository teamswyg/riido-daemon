//go:build !darwin

package securecredential

import "context"

// No portable keychain/DPAPI fallback substitutes for OS user presence.
func verifyHelper(context.Context, string) error { return ErrUnavailable }
