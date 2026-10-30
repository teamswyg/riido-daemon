package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func (s *Service) Repositories(ctx context.Context, id, teamID string) (*userapi.RepositoryList, error) {
	return s.reader.Repositories(ctx, id, teamID)
}
