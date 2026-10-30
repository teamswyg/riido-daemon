package standalonecli

import (
	"context"
	"errors"

	"github.com/teamswyg/riido-daemon/internal/userauth"
)

type Status struct {
	Mode              string `json:"mode"`
	APIOrigin         string `json:"api_origin"`
	CredentialPresent bool   `json:"credential_present"`
	UserCredential    string `json:"user_credential"`
	APIVerification   string `json:"api_verification"`
}

func (s *Service) Status(ctx context.Context) (Status, error) {
	present, err := s.store.Status(ctx)
	if err != nil {
		return Status{}, err
	}
	state := "absent"
	if present {
		state = "stored"
	}
	return Status{Mode: "user", APIOrigin: s.origin, CredentialPresent: present, UserCredential: state, APIVerification: "not_checked"}, nil
}

func (s *Service) Logout(ctx context.Context) error {
	err := s.manager.Logout(ctx)
	if err == nil {
		return nil
	}
	if errors.Is(err, userauth.ErrLocalRequestCleanupFailed) {
		return err
	}
	check, cancel := cleanupContext(ctx)
	defer cancel()
	if present, statusErr := s.store.Status(check); statusErr == nil && !present {
		return ErrLogoutRemoteIncomplete
	}
	return err
}
