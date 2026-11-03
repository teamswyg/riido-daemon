package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func (s *Service) Search(ctx context.Context, q userapi.WorkSearchRequest) (*userapi.WorkSearchPage, error) {
	return s.reader.Search(ctx, q)
}
