package standalonecli

import (
	"context"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func (s *Service) Relations(ctx context.Context, id, teamID string) ([]*userapi.WorkRelation, error) {
	return s.reader.Relations(ctx, id, teamID)
}
