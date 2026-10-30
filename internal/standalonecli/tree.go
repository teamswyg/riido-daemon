package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func (s *Service) Tree(ctx context.Context, id, teamID string) (*userapi.WorkTree, error) {
	return s.reader.Tree(ctx, id, teamID)
}
