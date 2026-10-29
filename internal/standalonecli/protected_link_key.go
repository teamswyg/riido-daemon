package standalonecli

import (
	"context"
	"errors"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

type protectedLinkKeyStore interface {
	ExistingInstallationKey(context.Context, string) (securecredential.InstallationPublicKey, error)
	SignDeviceLinkProof(context.Context, securecredential.InstallationPublicKey, securecredential.DeviceLinkProof) (string, error)
}

// These link-only adapters are not prerequisites for ordinary standalone login.
func (s *ProtectedStore) ExistingInstallationKey(ctx context.Context, resource string) (userauth.InstallationKey, error) {
	if s == nil || s.Native == nil || resource != s.Resource || !validLinkResource(resource) {
		return userauth.InstallationKey{}, securecredential.ErrUnavailable
	}
	native, ok := s.Native.(protectedLinkKeyStore)
	if !ok {
		return userauth.InstallationKey{}, securecredential.ErrUnavailable
	}
	key, err := native.ExistingInstallationKey(ctx, resource)
	return userauth.InstallationKey(key), mapLinkKeyError(err)
}

func (s *ProtectedStore) SignDeviceLinkProof(ctx context.Context, key userauth.InstallationKey, proof userauth.DeviceLinkProof) (string, error) {
	if s == nil || s.Native == nil || proof.Resource != s.Resource || !validLinkResource(proof.Resource) || proof.ClientID != "riido-cli" || proof.KeyFingerprint != key.KeyFingerprint {
		return "", securecredential.ErrUnavailable
	}
	native, ok := s.Native.(protectedLinkKeyStore)
	if !ok {
		return "", securecredential.ErrUnavailable
	}
	signature, err := native.SignDeviceLinkProof(ctx, securecredential.InstallationPublicKey(key), securecredential.DeviceLinkProof(proof))
	return signature, mapLinkKeyError(err)
}

func mapLinkKeyError(err error) error {
	if errors.Is(err, securecredential.ErrNotFound) {
		return securecredential.ErrNotFound
	}
	return mapStorageError(err)
}
