package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func (s *Service) Comments(ctx context.Context, id, teamID string, pagination userapi.CommentPagination) (userapi.CommentPage, error) {
	return s.reader.Comments(ctx, id, teamID, pagination)
}

func (s *Service) Replies(ctx context.Context, id, teamID, commentID string) ([]*userapi.Comment, error) {
	return s.reader.Replies(ctx, id, teamID, commentID)
}
