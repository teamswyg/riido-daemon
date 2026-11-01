package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func (s *Service) Agents(ctx context.Context, workspace, task, cursor string) (*userapi.AgentPage, error) {
	return (userapi.AgentReader{Client: s.manager}).List(ctx, workspace, task, cursor)
}

func (s *Service) Agent(ctx context.Context, workspace, task, id string) (*userapi.AgentDetail, error) {
	return (userapi.AgentReader{Client: s.manager}).Show(ctx, workspace, task, id)
}
