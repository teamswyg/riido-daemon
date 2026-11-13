package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/userapi"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

// Resolve and consume one current locator under the same protected read
// session. A canonical ID also resolves, so a stale team is never fabricated.
func readWorkspaceLocator[T any](ctx context.Context, s *Service, q userapi.WorkReferenceRequest, read func(userapi.Reader, userapi.WorkLocator) (T, error)) (T, error) {
	var result T
	if q.LocatorVersion != 0 && q.LocatorVersion != 2 {
		return result, userapi.ErrInvalidRequest
	}
	q.LocatorVersion = 2
	if !q.Valid() {
		return result, userapi.ErrInvalidRequest
	}
	err := s.manager.WithReadSession(ctx, func(client userauth.SessionReadClient) error {
		reader := userapi.Reader{Endpoint: s.reader.Endpoint, Client: client}
		resolved, err := reader.ResolveWorkReference(ctx, q)
		if err != nil {
			return err
		}
		result, err = read(reader, resolved.Locator())
		return err
	})
	if err != nil {
		var empty T
		return empty, err
	}
	return result, nil
}

func (s *Service) WorkV2(ctx context.Context, q userapi.WorkReferenceRequest) (userapi.WorkV2, error) {
	return readWorkspaceLocator(ctx, s, q, func(r userapi.Reader, locator userapi.WorkLocator) (userapi.WorkV2, error) {
		return r.WorkV2(ctx, locator)
	})
}

func (s *Service) ContextV2(ctx context.Context, q userapi.WorkReferenceRequest) (userapi.ReadScopeV2, error) {
	return readWorkspaceLocator(ctx, s, q, func(r userapi.Reader, locator userapi.WorkLocator) (userapi.ReadScopeV2, error) {
		return r.ContextV2(ctx, locator)
	})
}

func (s *Service) BodyV2(ctx context.Context, ref userapi.WorkReferenceRequest, q userapi.ReadBodyRequest) (userapi.ReadBodyPageV2, error) {
	if !q.ValidContent() || q.Kind == "meeting" {
		return userapi.ReadBodyPageV2{}, userapi.ErrInvalidRequest
	}
	return readWorkspaceLocator(ctx, s, ref, func(r userapi.Reader, locator userapi.WorkLocator) (userapi.ReadBodyPageV2, error) {
		q.ComponentID = locator.ID
		return r.BodyV2(ctx, locator, q)
	})
}

func (s *Service) ScopeCommentsV2(ctx context.Context, ref userapi.WorkReferenceRequest, q userapi.ReadScopeCommentsRequest) (userapi.ReadScopeCommentPageV2, error) {
	if !q.ValidPage() || q.ResourceKind == "meeting" {
		return userapi.ReadScopeCommentPageV2{}, userapi.ErrInvalidRequest
	}
	return readWorkspaceLocator(ctx, s, ref, func(r userapi.Reader, locator userapi.WorkLocator) (userapi.ReadScopeCommentPageV2, error) {
		q.ResourceID = locator.ID
		return r.ScopeCommentsV2(ctx, locator, q)
	})
}
