package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func (s *Service) Threads(ctx context.Context, workspace, task string) (*userapi.ThreadList, error) {
	return (userapi.ThreadReader{Client: s.manager}).List(ctx, workspace, task)
}

func (s *Service) Thread(ctx context.Context, workspace, task, id string) (*userapi.ThreadShow, error) {
	return (userapi.ThreadReader{Client: s.manager}).Show(ctx, workspace, task, id)
}
