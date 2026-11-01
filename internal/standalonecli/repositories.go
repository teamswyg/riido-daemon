package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/userapi"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func (s *Service) Repositories(ctx context.Context, id, teamID string) (*userapi.RepositoryList, error) {
	if id == "" || teamID == "" {
		return nil, userapi.ErrInvalidResponse
	}
	var result *userapi.RepositoryList
	err := s.manager.WithReadSession(ctx, func(client userauth.SessionReadClient) error {
		reader := userapi.Reader{Endpoint: s.reader.Endpoint, Client: client}
		var err error
		result, err = reader.Repositories(ctx, id, teamID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
