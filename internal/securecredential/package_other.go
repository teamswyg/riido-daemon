//go:build !darwin

package securecredential

import "context"

// VerifyPackagedHelper deliberately fails on hosts that cannot verify Apple code.
func VerifyPackagedHelper(context.Context, string, []string) error { return ErrUnavailable }
