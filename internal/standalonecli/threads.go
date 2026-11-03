package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/userapi"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func (s *Service) Threads(ctx context.Context, workspace, task string, team ...string) (*userapi.ThreadList, error) {
	return readCPWorkReference(ctx, s, workspace, task, team, userauth.ErrThreadReadUnsupported, func(ctx context.Context, client userauth.SessionReadClient, workspace, task string) (*userapi.ThreadList, error) {
		return (userapi.ThreadReader{Client: client}).List(ctx, workspace, task)
	})
}

func (s *Service) Thread(ctx context.Context, workspace, task, id string, team ...string) (*userapi.ThreadShow, error) {
	if !userapi.ValidThreadReadID(id) {
		return nil, userapi.ErrInvalidRequest
	}
	return readCPWorkReference(ctx, s, workspace, task, team, userauth.ErrThreadReadUnsupported, func(ctx context.Context, client userauth.SessionReadClient, workspace, task string) (*userapi.ThreadShow, error) {
		return (userapi.ThreadReader{Client: client}).Show(ctx, workspace, task, id)
	})
}
