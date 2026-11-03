package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/userapi"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func (s *Service) Workspaces(ctx context.Context, q userapi.WorkspaceListRequest) (*userapi.WorkspaceList, error) {
	var result *userapi.WorkspaceList
	err := s.manager.WithReadSession(ctx, func(client userauth.SessionReadClient) error {
		reader := userapi.Reader{Endpoint: s.reader.Endpoint, Client: client}
		var err error
		result, err = reader.Workspaces(ctx, q)
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Workspace confirms membership within one protected login session. The caller
// may persist a preference only after WithReadSession's final checks succeed.
func (s *Service) Workspace(ctx context.Context, id string) (*userapi.WorkspaceItem, error) {
	var result *userapi.WorkspaceItem
	err := s.manager.WithReadSession(ctx, func(client userauth.SessionReadClient) error {
		reader := userapi.Reader{Endpoint: s.reader.Endpoint, Client: client}
		q := userapi.WorkspaceListRequest{Limit: 100}
		seen := map[string]bool{}
		userID, previous := "", ""
		for {
			page, err := reader.Workspaces(ctx, q)
			if err != nil {
				return err
			}
			if userID != "" && page.UserID != userID {
				return userapi.ErrInvalidResponse
			}
			userID = page.UserID
			for _, item := range page.Items {
				if item.ID <= previous {
					return userapi.ErrInvalidResponse
				}
				previous = item.ID
				if item.ID == id {
					selected := item
					result = &selected
					return nil
				}
			}
			if page.Complete {
				return userapi.ErrPermission
			}
			if seen[*page.NextCursor] {
				return userapi.ErrInvalidResponse
			}
			seen[*page.NextCursor] = true
			q.Cursor = page.NextCursor
		}
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
