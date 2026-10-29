package standalonecli

import (
	"context"
	"errors"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

// protectedNativeStore is implemented by the verified native OS helper. There
// is no file or environment credential implementation in the CLI assembly.
type protectedNativeStore interface {
	Lock(context.Context) (func(), error)
	Status(context.Context) (bool, error)
	ConfirmUserPresence(context.Context) error
	EnsureInstallationKey(context.Context, string) (securecredential.InstallationPublicKey, error)
	SignInstallationProof(context.Context, securecredential.InstallationPublicKey, securecredential.InstallationProof) (string, error)
	Load(context.Context) ([]byte, error)
	Save(context.Context, []byte) error
	Delete(context.Context) error
}

type ProtectedStore struct {
	Native   protectedNativeStore
	Resource string
}

func (s *ProtectedStore) Lock(ctx context.Context) (func(), error) {
	if s == nil || s.Native == nil {
		return nil, securecredential.ErrUnavailable
	}
	unlock, err := s.Native.Lock(ctx)
	if err == nil && unlock == nil {
		return nil, securecredential.ErrUnavailable
	}
	return unlock, mapStorageError(err)
}

func (s *ProtectedStore) Status(ctx context.Context) (bool, error) {
	present, err := s.Native.Status(ctx)
	return present, mapStorageError(err)
}

func (s *ProtectedStore) ConfirmUserPresence(ctx context.Context) error {
	return mapStorageError(s.Native.ConfirmUserPresence(ctx))
}

func (s *ProtectedStore) EnsureInstallationKey(ctx context.Context, resource string) (userauth.InstallationKey, error) {
	if s == nil || s.Native == nil {
		return userauth.InstallationKey{}, securecredential.ErrUnavailable
	}
	key, err := s.Native.EnsureInstallationKey(ctx, resource)
	return userauth.InstallationKey(key), mapStorageError(err)
}

func (s *ProtectedStore) SignInstallationProof(ctx context.Context, key userauth.InstallationKey, proof userauth.InstallationProof) (string, error) {
	if s == nil || s.Native == nil {
		return "", securecredential.ErrUnavailable
	}
	signature, err := s.Native.SignInstallationProof(ctx, securecredential.InstallationPublicKey(key), securecredential.InstallationProof(proof))
	return signature, mapStorageError(err)
}

func (s *ProtectedStore) Load(ctx context.Context) (userauth.Credential, error) {
	record, err := s.loadRecord(ctx)
	return record.Credential, err
}

// Save is called under CredentialStore.Lock by the authentication manager.
// Reloading under that same lock preserves pending state across token rotation.
func (s *ProtectedStore) Save(ctx context.Context, credential userauth.Credential) error {
	record, err := s.loadRecord(ctx)
	if err != nil && !errors.Is(err, userauth.ErrNotAuthenticated) {
		return err
	}
	if err != nil || record.Credential.SessionID != credential.SessionID {
		record = protectedSession{SchemaVersion: protectedSessionVersion}
	}
	record.Credential = credential
	return s.saveRecord(ctx, record)
}

func (s *ProtectedStore) loadRecord(ctx context.Context) (protectedSession, error) {
	if s == nil || s.Native == nil {
		return protectedSession{}, securecredential.ErrUnavailable
	}
	raw, err := s.Native.Load(ctx)
	if err != nil {
		return protectedSession{}, mapStorageError(err)
	}
	record, err := decodeProtectedSession(raw)
	if err == nil && record.PendingLink != nil && record.PendingLink.Resource != s.Resource {
		return protectedSession{}, securecredential.ErrUnavailable
	}
	return record, err
}

func (s *ProtectedStore) saveRecord(ctx context.Context, record protectedSession) error {
	raw, err := encodeProtectedSession(record)
	if err != nil {
		return err
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
