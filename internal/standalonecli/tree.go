package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func (s *Service) Tree(ctx context.Context, id, teamID string, workspace ...string) (*userapi.WorkTree, error) {
	return readWorkReference(ctx, s, id, teamID, workspace, false, func(r userapi.Reader, id, team string) (*userapi.WorkTree, error) { return r.Tree(ctx, id, team) })
}
