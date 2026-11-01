package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/userapi"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func (s *Service) Workspaces(ctx context.Context, q userapi.WorkspaceListRequest) (*userapi.WorkspaceList, error) {
	var result *userapi.WorkspaceList
	err := s.manager.WithReadSession(ctx, func(client userauth.SessionReadClient) error {
		reader := userapi.Reader{Endpoint: s.reader.Endpoint, Client: client}
		var err error
		result, err = reader.Workspaces(ctx, q)
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
