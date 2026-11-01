package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/userapi"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func (s *Service) Agents(ctx context.Context, workspace, task, cursor string, team ...string) (*userapi.AgentPage, error) {
	if !userapi.ValidAgentCursor(cursor) {
		return nil, userapi.ErrInvalidRequest
	}
	return readCPWorkReference(ctx, s, workspace, task, team, userauth.ErrAgentReadUnsupported, func(ctx context.Context, client userauth.SessionReadClient, workspace, task string) (*userapi.AgentPage, error) {
		return (userapi.AgentReader{Client: client}).List(ctx, workspace, task, cursor)
	})
}

func (s *Service) Agent(ctx context.Context, workspace, task, id string, team ...string) (*userapi.AgentDetail, error) {
	if !userapi.ValidThreadReadID(id) {
		return nil, userapi.ErrInvalidRequest
	}
	return readCPWorkReference(ctx, s, workspace, task, team, userauth.ErrAgentReadUnsupported, func(ctx context.Context, client userauth.SessionReadClient, workspace, task string) (*userapi.AgentDetail, error) {
		return (userapi.AgentReader{Client: client}).Show(ctx, workspace, task, id)
	})
}
