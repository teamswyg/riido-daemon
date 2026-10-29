package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func (s *Service) Activity(ctx context.Context, id, teamID string, pagination userapi.ActivityPagination) (userapi.ActivityPage, error) {
	return s.reader.Activity(ctx, id, teamID, pagination)
}
