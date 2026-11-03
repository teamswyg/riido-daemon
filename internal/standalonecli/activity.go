package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func (s *Service) Activity(ctx context.Context, id, teamID string, pagination userapi.ActivityPagination, workspace ...string) (userapi.ActivityPage, error) {
	if !pagination.Valid() {
		return userapi.ActivityPage{}, userapi.ErrInvalidResponse
	}
	return readWorkReference(ctx, s, id, teamID, workspace, false, func(r userapi.Reader, id, team string) (userapi.ActivityPage, error) {
		return r.Activity(ctx, id, team, pagination)
	})
}
