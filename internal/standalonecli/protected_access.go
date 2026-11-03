package standalonecli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

// protectedAccessNativeStore is the optional helper item for the access grant
// (RIID-6014). It is this-device-only keychain storage without a user-presence
// prompt and never holds refresh tokens, recovery secrets or keys.
type protectedAccessNativeStore interface {
	LoadAccess(context.Context) ([]byte, error)
	SaveAccess(context.Context, []byte) error
	DeleteAccess(context.Context) error
}

const protectedAccessVersion = "riido-cli-access.v1"

type protectedAccessRecord struct {
	SchemaVersion string               `json:"schema_version"`
	Grant         userauth.AccessGrant `json:"grant"`
}

func (s *ProtectedStore) accessNative() (protectedAccessNativeStore, bool) {
	if s == nil || s.Native == nil {
		return nil, false
	}
	native, ok := s.Native.(protectedAccessNativeStore)
	return native, ok
}

// LoadAccess implements userauth.AccessCache. Any decoding or binding doubt is
// reported as unavailable so the manager falls back to the protected credential.
func (s *ProtectedStore) LoadAccess(ctx context.Context) (userauth.AccessGrant, error) {
	native, ok := s.accessNative()
	if !ok {
		return userauth.AccessGrant{}, securecredential.ErrUnavailable
	}
	raw, err := native.LoadAccess(ctx)
	if err != nil {
		return userauth.AccessGrant{}, mapStorageError(err)
	}
	var record protectedAccessRecord
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&record) != nil || decoder.Decode(new(any)) != io.EOF || record.SchemaVersion != protectedAccessVersion {
		return userauth.AccessGrant{}, securecredential.ErrUnavailable
	}
	if !s.acceptsCredential(userauth.Credential{Binding: record.Grant.Binding}) {
		return userauth.AccessGrant{}, securecredential.ErrUnavailable
	}
	return record.Grant, nil
}

// SaveAccess implements userauth.AccessCache.
func (s *ProtectedStore) SaveAccess(ctx context.Context, grant userauth.AccessGrant) error {
	native, ok := s.accessNative()
	if !ok || grant.SessionID == "" || grant.AccessToken == "" || !s.acceptsCredential(userauth.Credential{Binding: grant.Binding}) {
		return securecredential.ErrUnavailable
	}
	raw, err := json.Marshal(protectedAccessRecord{SchemaVersion: protectedAccessVersion, Grant: grant})
	if err != nil {
		return securecredential.ErrUnavailable
	}
	return mapStorageError(native.SaveAccess(ctx, raw))
}

// syncAccess follows a protected Save: the cache holds the saved grant, or
// nothing while a refresh rotation is pending. If the cache cannot be written
// it is cleared; a cache that can be neither written nor cleared is still only
// an access token that the server rejects once rotated.
func (s *ProtectedStore) syncAccess(ctx context.Context, credential userauth.Credential) {
	native, ok := s.accessNative()
	if !ok {
		return
	}
	if credential.RefreshRecoverySecret == "" && s.SaveAccess(ctx, credential.AccessGrant()) == nil {
		return
	}
	_ = native.DeleteAccess(ctx)
}

// clearAccess runs before the protected credential is saved or deleted, so a
// replaced or logged-out access token never stays readable without user
// presence. A helper without
// the access cache has nothing to clear.
func (s *ProtectedStore) clearAccess(ctx context.Context) error {
	native, ok := s.accessNative()
	if !ok {
		return nil
	}
	err := native.DeleteAccess(ctx)
	if err == nil || errors.Is(err, securecredential.ErrNotFound) || errors.Is(err, securecredential.ErrUnsupported) {
		return nil
	}
	return mapStorageError(err)
}
