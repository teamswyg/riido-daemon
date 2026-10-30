package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func (s *Service) Resources(ctx context.Context, id, teamID string) (*userapi.Resources, error) {
	return s.reader.Resources(ctx, id, teamID)
}
