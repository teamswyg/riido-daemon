package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func (s *Service) Repositories(ctx context.Context, id, teamID string, workspace ...string) (*userapi.RepositoryList, error) {
	if id == "" || (teamID == "" && len(workspace) == 0) {
		return nil, userapi.ErrInvalidResponse
	}
	return readWorkReference(ctx, s, id, teamID, workspace, true, func(r userapi.Reader, id, team string) (*userapi.RepositoryList, error) {
		return r.Repositories(ctx, id, team)
	})
}
