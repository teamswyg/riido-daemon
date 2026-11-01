package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func (s *Service) Comments(ctx context.Context, id, teamID string, pagination userapi.CommentPagination, workspace ...string) (userapi.CommentPage, error) {
	if !pagination.Valid() {
		return userapi.CommentPage{}, userapi.ErrInvalidResponse
	}
	return readWorkReference(ctx, s, id, teamID, workspace, false, func(r userapi.Reader, id, team string) (userapi.CommentPage, error) {
		return r.Comments(ctx, id, team, pagination)
	})
}

func (s *Service) Replies(ctx context.Context, id, teamID, commentID string, workspace ...string) ([]*userapi.Comment, error) {
	if commentID == "" {
		return nil, userapi.ErrInvalidResponse
	}
	return readWorkReference(ctx, s, id, teamID, workspace, false, func(r userapi.Reader, id, team string) ([]*userapi.Comment, error) {
		return r.Replies(ctx, id, team, commentID)
	})
}
