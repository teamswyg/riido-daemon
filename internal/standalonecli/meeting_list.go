package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/userapi"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func (s *Service) Meetings(ctx context.Context, workID, teamID string) (*userapi.MeetingList, error) {
	var result *userapi.MeetingList
	err := s.manager.WithReadSession(ctx, func(client userauth.SessionReadClient) error {
		reader := userapi.Reader{Endpoint: s.reader.Endpoint, Client: client}
		var err error
		result, err = reader.Meetings(ctx, workID, teamID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
