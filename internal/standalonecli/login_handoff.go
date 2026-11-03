package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func (s *Service) beginLogin(ctx context.Context, key userauth.InstallationKey) (*userauth.Authorization, error) {
	sign := func(ctx context.Context, proof userauth.InstallationProof) (string, error) {
		return s.store.SignInstallationProof(ctx, key, proof)
	}
	var flow *userauth.Authorization
	var err error
	var open func() error
	openErr := ErrBrowserUnavailable
	if s.authOrigin == userauth.AuthOriginDesktop {
		if s.openDesktop == nil {
			return nil, ErrDesktopUnavailable
		}
		flow, err = s.client.BeginDesktopAuthorization(ctx, key, sign)
		openErr = ErrDesktopUnavailable
		open = func() error {
			request, err := flow.DesktopRequest()
			if err != nil {
				return err
			}
			return s.openDesktop(ctx, request)
		}
	} else {
		if s.openBrowser == nil {
			return nil, ErrBrowserUnavailable
		}
		flow, err = s.client.BeginInstallationAuthorization(ctx, key, sign)
		open = func() error { return s.openBrowser(ctx, flow.URL) }
	}
	if err != nil {
		return nil, err
	}
	if err := open(); err != nil {
		flow.Close()
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, openErr
	}
	return flow, nil
}
