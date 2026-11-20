package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

// Desktop invokes the installed CLI through private stdio. The CLI retains the
// OAuth credential and private key; the Desktop retains both native secrets.
func (s *Service) DesktopBootstrap(ctx context.Context, op string, request []byte) (any, error) {
	var input userauth.BootstrapRequest
	switch op {
	case "identity":
		if request != nil {
			return nil, linkError("BOOTSTRAP_INVALID_REQUEST")
		}
	case "request", "restart":
		var err error
		input, err = userauth.ParseBootstrapRequest(request, op == "restart")
		if err != nil {
			return nil, err
		}
	default:
		return nil, linkError("BOOTSTRAP_INVALID_REQUEST")
	}
	if op == "identity" {
		return s.bootstrapIdentity(ctx)
	}
	return s.BoundBootstrapRequest(ctx, input)
}

// BoundBootstrapRequest compares all installation and account fields with the currently protected session.
func (s *Service) BoundBootstrapRequest(ctx context.Context, input userauth.BootstrapRequest) (userauth.DeviceLinkApproval, error) {
	id, session, err := s.currentBootstrapIdentity(ctx)
	if err != nil {
		return userauth.DeviceLinkApproval{}, err
	}
	if input.BootstrapIdentity != id {
		return userauth.DeviceLinkApproval{}, linkError("LINK_BINDING_MISMATCH")
	}
	return s.manager.RequestBootstrap(ctx, session, input)
}

func (s *Service) bootstrapIdentity(ctx context.Context) (userauth.BootstrapIdentity, error) {
	id, _, err := s.currentBootstrapIdentity(ctx)
	return id, err
}

func (s *Service) currentBootstrapIdentity(ctx context.Context) (userauth.BootstrapIdentity, string, error) {
	store, ok := s.store.(*ProtectedStore)
	if !ok {
		return userauth.BootstrapIdentity{}, "", securecredential.ErrUnavailable
	}
	session, err := s.linkSession(ctx)
	if err != nil {
		return userauth.BootstrapIdentity{}, "", err
	}
	key, err := store.ExistingInstallationKey(ctx, s.origin)
	if err != nil {
		return userauth.BootstrapIdentity{}, "", err
	}
	id, err := s.manager.BootstrapIdentity(ctx, session)
	if err != nil {
		return userauth.BootstrapIdentity{}, "", err
	}
	if id.KeyFingerprint != key.KeyFingerprint {
		return userauth.BootstrapIdentity{}, "", linkError("LINK_BINDING_MISMATCH")
	}
	return id, session, nil
}
