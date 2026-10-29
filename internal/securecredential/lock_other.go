//go:build !darwin && !linux

package securecredential

import "context"

// User mode remains closed on platforms lacking the full protected store.
func (s *Store) Lock(context.Context) (func(), error) { return nil, ErrUnavailable }
