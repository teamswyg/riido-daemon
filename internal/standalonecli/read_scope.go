package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func (s *Service) Context(ctx context.Context, id, teamID string) (userapi.ReadScope, error) {
	return s.reader.Context(ctx, id, teamID)
}

func (s *Service) Body(ctx context.Context, q userapi.ReadBodyRequest) (userapi.ReadBodyPage, error) {
	return s.reader.Body(ctx, q)
}

func (s *Service) Meeting(ctx context.Context, q userapi.MeetingReadRequest) (userapi.ReadMeeting, error) {
	return s.reader.Meeting(ctx, q)
}

func (s *Service) ScopeComments(ctx context.Context, q userapi.ReadScopeCommentsRequest) (userapi.ReadScopeCommentPage, error) {
	return s.reader.ScopeComments(ctx, q)
}
