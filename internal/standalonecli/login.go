package standalonecli

import (
	"context"
	"errors"
	"time"

	"github.com/teamswyg/riido-daemon/internal/userapi"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

type LoginResult struct {
	Mode          string           `json:"mode"`
	Authenticated bool             `json:"authenticated"`
	APIOrigin     string           `json:"api_origin"`
	Identity      userapi.Identity `json:"identity"`
	Warnings      []string         `json:"warnings,omitempty"`
}

func (s *Service) Login(ctx context.Context) (LoginResult, error) {
	if err := ctx.Err(); err != nil {
		return LoginResult{}, err
	}
	if _, err := s.store.Status(ctx); err != nil {
		return LoginResult{}, err
	}
	if err := s.store.ConfirmUserPresence(ctx); err != nil {
		return LoginResult{}, err
	}
	if err := ctx.Err(); err != nil {
		return LoginResult{}, err
	}
	if s.openBrowser == nil {
		return LoginResult{}, ErrBrowserUnavailable
	}
	flow, err := s.client.BeginAuthorization(ctx)
	if err != nil {
		return LoginResult{}, err
	}
	defer flow.Close()
	if err := s.openBrowser(ctx, flow.URL); err != nil {
		if ctx.Err() != nil {
			return LoginResult{}, ctx.Err()
		}
		return LoginResult{}, ErrBrowserUnavailable
	}
	credential, err := flow.Wait(ctx)
	if err != nil {
		return LoginResult{}, err
	}
	identity, verified, err := s.verifyNewGrant(ctx, credential)
	if err != nil {
		s.cleanupNewGrant(ctx, verified)
		return LoginResult{}, err
	}
	credential = verified
	loginErr := s.manager.Login(ctx, credential)
	partial := errors.Is(loginErr, userauth.ErrPreviousGrantRevocationFailed)
	if loginErr != nil && !partial {
		s.cleanupNewGrant(ctx, credential)
		return LoginResult{}, loginErr
	}
	result := LoginResult{Mode: "user", Authenticated: true, APIOrigin: s.origin, Identity: identity}
	if partial {
		result.Warnings = []string{"previous_grant_revocation_failed"}
	}
	return result, nil
}

func (s *Service) cleanupNewGrant(ctx context.Context, credential userauth.Credential) {
	cleanup, cancel := cleanupContext(ctx)
	defer cancel()
	token := credential.RefreshToken
	if token == "" {
		token = credential.AccessToken
	}
	_ = s.client.Revoke(cleanup, token)
}

func cleanupContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
}
