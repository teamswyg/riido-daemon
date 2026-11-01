package standalonecli

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/userapi"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

// Resolve keys and consume a fixed-purpose CP response under one login. This
// scope includes DTO parsing and discards results if the final login check fails.
func readCPWorkReference[T any](ctx context.Context, s *Service, workspace, task string, team []string, unsupported error, read func(context.Context, userauth.SessionReadClient, string, string) (T, error)) (T, error) {
	var empty T
	if !s.manager.SupportsTaskReads() {
		return empty, unsupported
	}
	if len(team) > 1 {
		return empty, userapi.ErrInvalidRequest
	}
	ref, parseErr := contextcli.ParseWorkRef(task)
	isKey := parseErr == nil && ref.Kind == contextcli.WorkRefKey
	q := userapi.WorkReferenceRequest{Reference: ref.Value, WorkspaceID: workspace}
	if len(team) == 1 {
		q.TeamID = team[0]
	}
	if isKey {
		if !q.Valid() {
			return empty, userapi.ErrInvalidRequest
		}
	} else if !userapi.ValidThreadReadID(workspace) || !userapi.ValidThreadReadID(task) || q.TeamID != "" {
		return empty, userapi.ErrInvalidRequest
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	var result T
	err := s.manager.WithReadSession(ctx, func(client userauth.SessionReadClient) error {
		if isKey {
			target, err := (userapi.Reader{Endpoint: s.reader.Endpoint, Client: client}).ResolveWorkReference(ctx, q)
			if err != nil {
				return err
			}
			workspace, task = target.WorkspaceID, target.ID
		}
		var err error
		result, err = read(ctx, client, workspace, task)
		return err
	})
	if err != nil {
		return empty, err
	}
	return result, nil
}
