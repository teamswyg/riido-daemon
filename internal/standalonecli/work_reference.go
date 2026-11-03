package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/userapi"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

// readWorkReference preserves the existing opaque-ID route. Only keys add a
// lookup. Compound readers hold the same session even when given an ID.
func readWorkReference[T any](ctx context.Context, s *Service, id, team string, workspace []string, compound bool, read func(userapi.Reader, string, string) (T, error)) (T, error) {
	var empty T
	if len(workspace) > 1 {
		return empty, userapi.ErrInvalidRequest
	}
	ref, parseErr := contextcli.ParseWorkRef(id)
	isKey := parseErr == nil && ref.Kind == contextcli.WorkRefKey
	if compound && !isKey && (id == "" || team == "") {
		return empty, userapi.ErrInvalidResponse
	}
	if !isKey && !compound {
		return read(s.reader, id, team)
	}
	q := userapi.WorkReferenceRequest{Reference: ref.Value, TeamID: team}
	if len(workspace) == 1 {
		q.WorkspaceID = workspace[0]
	}
	if isKey && !q.Valid() {
		return empty, userapi.ErrInvalidRequest
	}
	var result T
	err := s.manager.WithReadSession(ctx, func(client userauth.SessionReadClient) error {
		reader := userapi.Reader{Endpoint: s.reader.Endpoint, Client: client}
		if isKey {
			target, err := reader.ResolveWorkReference(ctx, q)
			if err != nil {
				return err
			}
			id, team = target.ID, target.TeamID
		}
		var err error
		result, err = read(reader, id, team)
		return err
	})
	if err != nil {
		return empty, err
	}
	return result, nil
}
