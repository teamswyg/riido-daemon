package standalonecli

import (
	"context"
	"sync"

	"github.com/teamswyg/riido-daemon/internal/userapi"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

// A fresh OAuth grant is verified before replacing a protected stored session.
// This in-memory store is private to one login attempt and never reads existing
// credentials. If the API rotates the grant, cleanup uses the rotated value.
func (s *Service) verifyNewGrant(ctx context.Context, credential userauth.Credential) (userapi.Identity, userauth.Credential, error) {
	pending := &pendingGrant{}
	manager := userauth.NewManager(s.client, pending)
	if err := manager.Login(ctx, credential); err != nil {
		return userapi.Identity{}, credential, err
	}
	reader := userapi.Reader{Endpoint: s.origin + "/graphql", Client: manager}
	identity, err := reader.Whoami(ctx)
	return identity, pending.credential, err
}

type pendingGrant struct {
	mu         sync.Mutex
	credential userauth.Credential
	present    bool
}

func (p *pendingGrant) Lock(context.Context) (func(), error) { p.mu.Lock(); return p.mu.Unlock, nil }
func (p *pendingGrant) Load(context.Context) (userauth.Credential, error) {
	if !p.present {
		return userauth.Credential{}, userauth.ErrNotAuthenticated
	}
	return p.credential, nil
}

func (p *pendingGrant) Save(_ context.Context, credential userauth.Credential) error {
	p.credential = credential
	p.present = true
	return nil
}
func (p *pendingGrant) Delete(context.Context) error { p.present = false; return nil }
