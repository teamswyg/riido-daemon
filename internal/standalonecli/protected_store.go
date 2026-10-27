package standalonecli

import (
	"context"
	"errors"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

type ProtectedStore struct{ Native *securecredential.Store }

func (s *ProtectedStore) Lock(ctx context.Context) (func(), error) {
	unlock, err := s.Native.Lock(ctx)
	return unlock, mapStorageError(err)
}

func (s *ProtectedStore) Status(ctx context.Context) (bool, error) {
	present, err := s.Native.Status(ctx)
	return present, mapStorageError(err)
}

func (s *ProtectedStore) ConfirmUserPresence(ctx context.Context) error {
	return mapStorageError(s.Native.ConfirmUserPresence(ctx))
}

func (s *ProtectedStore) Load(ctx context.Context) (userauth.Credential, error) {
	var result userauth.Credential
	raw, err := s.Native.Load(ctx)
	if err != nil {
		return result, mapStorageError(err)
	}
	return decodeCredential(raw)
}

func (s *ProtectedStore) Save(ctx context.Context, credential userauth.Credential) error {
	raw, err := encodeCredential(credential)
	if err != nil {
		return securecredential.ErrUnavailable
	}
	return mapStorageError(s.Native.Save(ctx, raw))
}

func (s *ProtectedStore) Delete(ctx context.Context) error {
	return mapStorageError(s.Native.Delete(ctx))
}

func mapStorageError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, securecredential.ErrNotFound):
		return userauth.ErrNotAuthenticated
	case errors.Is(err, securecredential.ErrInteractionRequired), errors.Is(err, securecredential.ErrCancelled):
		return userauth.ErrStoreLocked
	case errors.Is(err, context.Canceled):
		return context.Canceled
	case errors.Is(err, context.DeadlineExceeded):
		return context.DeadlineExceeded
	default:
		return securecredential.ErrUnavailable
	}
}
