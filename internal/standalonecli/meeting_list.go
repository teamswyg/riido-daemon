package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func (s *Service) Meetings(ctx context.Context, id, teamID string, workspace ...string) (*userapi.MeetingList, error) {
	return readWorkReference(ctx, s, id, teamID, workspace, true, func(r userapi.Reader, id, team string) (*userapi.MeetingList, error) {
		return r.Meetings(ctx, id, team)
	})
}
