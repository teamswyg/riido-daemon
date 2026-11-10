package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

// Desktop invokes the installed CLI through private stdio. The CLI retains the
// OAuth credential and private key; the Desktop retains both native secrets.
func (s *Service) DesktopBootstrap(ctx context.Context, request []byte) (any, error) {
	var input userauth.BootstrapRequest
	if request != nil {
		var err error
		input, err = userauth.ParseBootstrapRequest(request)
		if err != nil {
			return nil, err
		}
	}
	store, ok := s.store.(*ProtectedStore)
	if !ok {
		return nil, securecredential.ErrUnavailable
	}
	session, err := s.linkSession(ctx)
	if err != nil {
		return nil, err
	}
	key, err := store.ExistingInstallationKey(ctx, s.origin)
	if err != nil {
		return nil, err
	}
	id, err := s.manager.BootstrapIdentity(ctx, session)
	if err != nil {
		return nil, err
	}
	if id.KeyFingerprint != key.KeyFingerprint {
		return nil, linkError("LINK_BINDING_MISMATCH")
	}
	if request == nil {
		return id, nil
	}
	if input.BootstrapIdentity != id {
		return nil, linkError("LINK_BINDING_MISMATCH")
	}
	return s.manager.RequestBootstrap(ctx, session, input)
}
