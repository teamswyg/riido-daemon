package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func (s *Service) Relations(ctx context.Context, id, teamID string, workspace ...string) ([]*userapi.WorkRelation, error) {
	return readWorkReference(ctx, s, id, teamID, workspace, false, func(r userapi.Reader, id, team string) ([]*userapi.WorkRelation, error) {
		return r.Relations(ctx, id, team)
	})
}
