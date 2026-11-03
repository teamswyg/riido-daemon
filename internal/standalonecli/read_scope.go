package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func (s *Service) Context(ctx context.Context, id, teamID string, workspace ...string) (userapi.ReadScope, error) {
	if id == "" || (teamID == "" && len(workspace) == 0) {
		return userapi.ReadScope{}, userapi.ErrInvalidResponse
	}
	return readWorkReference(ctx, s, id, teamID, workspace, true, func(r userapi.Reader, id, team string) (userapi.ReadScope, error) { return r.Context(ctx, id, team) })
}

func (s *Service) Body(ctx context.Context, q userapi.ReadBodyRequest, workspace ...string) (userapi.ReadBodyPage, error) {
	if q.Kind != "" && q.Kind != "component" {
		return s.reader.Body(ctx, q)
	}
	if !q.ValidContent() {
		return userapi.ReadBodyPage{}, userapi.ErrInvalidResponse
	}
	return readWorkReference(ctx, s, q.ComponentID, q.TeamID, workspace, false, func(r userapi.Reader, id, team string) (userapi.ReadBodyPage, error) {
		q.ComponentID, q.TeamID = id, team
		return r.Body(ctx, q)
	})
}

func (s *Service) Meeting(ctx context.Context, q userapi.MeetingReadRequest) (userapi.ReadMeeting, error) {
	return s.reader.Meeting(ctx, q)
}

func (s *Service) ScopeComments(ctx context.Context, q userapi.ReadScopeCommentsRequest, workspace ...string) (userapi.ReadScopeCommentPage, error) {
	if q.ResourceKind != "component" {
		return s.reader.ScopeComments(ctx, q)
	}
	if !q.ValidPage() {
		return userapi.ReadScopeCommentPage{}, userapi.ErrInvalidResponse
	}
	return readWorkReference(ctx, s, q.ResourceID, q.TeamID, workspace, false, func(r userapi.Reader, id, team string) (userapi.ReadScopeCommentPage, error) {
		q.ResourceID, q.TeamID = id, team
		return r.ScopeComments(ctx, q)
	})
}
