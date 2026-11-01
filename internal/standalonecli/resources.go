package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func (s *Service) Resources(ctx context.Context, id, teamID string, workspace ...string) (*userapi.Resources, error) {
	if id == "" || (teamID == "" && len(workspace) == 0) {
		return nil, userapi.ErrInvalidResponse
	}
	return readWorkReference(ctx, s, id, teamID, workspace, true, func(r userapi.Reader, id, team string) (*userapi.Resources, error) { return r.Resources(ctx, id, team) })
}
