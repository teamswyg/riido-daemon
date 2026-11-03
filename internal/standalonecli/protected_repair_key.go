package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func (s *ProtectedStore) SignInstallationRepairProof(ctx context.Context, key userauth.InstallationKey, proof userauth.InstallationRepairProof) (string, error) {
	if s == nil || s.Native == nil || proof.Resource != s.Resource || !validLinkResource(proof.Resource) || proof.ClientID != "riido-cli" || proof.KeyFingerprint != key.KeyFingerprint {
		return "", securecredential.ErrUnavailable
	}
	native, ok := s.Native.(interface {
		SignInstallationRepairProof(context.Context, securecredential.InstallationPublicKey, securecredential.InstallationRepairProof) (string, error)
	})
	if !ok {
		return "", securecredential.ErrUnavailable
	}
	signature, err := native.SignInstallationRepairProof(ctx, securecredential.InstallationPublicKey(key), securecredential.InstallationRepairProof(proof))
	return signature, mapLinkKeyError(err)
}
