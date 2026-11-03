package securecredential

import (
	"context"
	"encoding/base64"
	"errors"
)

// maxAccessBytes bounds the cached access grant record.
const maxAccessBytes = 8 << 10

// ErrUnsupported means the installed helper predates the access cache
// (RIID-6014). Callers fall back to the user-presence protected credential.
var ErrUnsupported = errors.New("protected credential helper does not support the access cache")

// LoadAccess reads the short-lived access grant kept without user presence.
// Refresh tokens, recovery secrets and installation keys never use this item.
func (s *Store) LoadAccess(ctx context.Context) ([]byte, error) {
	r, err := s.callAccess(ctx, "access-load", nil)
	if err != nil {
		return nil, err
	}
	value, err := base64.StdEncoding.DecodeString(*r.Data.Value)
	if err != nil || len(value) == 0 || len(value) > maxAccessBytes {
		return nil, ErrUnavailable
	}
	return value, nil
}

func (s *Store) SaveAccess(ctx context.Context, value []byte) error {
	if len(value) == 0 || len(value) > maxAccessBytes {
		return ErrUnavailable
	}
	_, err := s.callAccess(ctx, "access-save", value)
	return err
}

func (s *Store) DeleteAccess(ctx context.Context) error {
	_, err := s.callAccess(ctx, "access-delete", nil)
	return err
}

func (s *Store) callAccess(ctx context.Context, op string, value []byte) (response, error) {
	r, err := s.call(ctx, op, value)
	if err != nil && r.Error != nil && r.Error.Code == "INVALID_REQUEST" {
		return r, ErrUnsupported
	}
	return r, err
}
