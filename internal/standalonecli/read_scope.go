package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/userapi"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func (s *Service) Context(ctx context.Context, id, teamID string) (userapi.ReadScope, error) {
	if id == "" || teamID == "" {
		return userapi.ReadScope{}, userapi.ErrInvalidResponse
	}
	var result userapi.ReadScope
	err := s.manager.WithReadSession(ctx, func(client userauth.SessionReadClient) error {
		reader := userapi.Reader{Endpoint: s.reader.Endpoint, Client: client}
		var err error
		result, err = reader.Context(ctx, id, teamID)
		return err
	})
	if err != nil {
		return userapi.ReadScope{}, err
	}
	return result, nil
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
