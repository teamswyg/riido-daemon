// Package standalonecli connects protected CLI authentication to user APIs.
package standalonecli

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/userapi"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

var (
	ErrBrowserUnavailable     = errors.New("could not open the authorization browser")
	ErrLogoutRemoteIncomplete = errors.New("local credentials were removed but server revocation was not confirmed")
)

type Store interface {
	userauth.CredentialStore
	Status(context.Context) (bool, error)
}
type Config struct {
	APIOrigin   string
	Store       Store
	OpenBrowser func(context.Context, string) error
	HTTPClient  *http.Client
}
type Service struct {
	client      *userauth.Client
	manager     *userauth.Manager
	reader      userapi.Reader
	store       Store
	origin      string
	openBrowser func(context.Context, string) error
}

func New(config Config) (*Service, error) {
	origin := strings.TrimRight(config.APIOrigin, "/")
	client, err := userauth.NewClient(userauth.Config{ClientID: "riido-cli", Resource: origin, Scopes: []string{"read"}, HTTPClient: config.HTTPClient, Endpoints: userauth.Endpoints{AuthorizeURL: origin + "/oauth/authorize", TokenURL: origin + "/oauth/token", RevokeURL: origin + "/oauth/revoke"}})
	if err != nil {
		return nil, err
	}
	if config.Store == nil {
		return nil, userauth.ErrStoreLocked
	}
	manager := userauth.NewManager(client, config.Store)
	return &Service{client: client, manager: manager, reader: userapi.Reader{Endpoint: origin + "/graphql", Client: manager}, store: config.Store, origin: origin, openBrowser: config.OpenBrowser}, nil
}

func (s *Service) Whoami(ctx context.Context) (userapi.Identity, error) { return s.reader.Whoami(ctx) }
func (s *Service) Work(ctx context.Context, id, teamID string) (userapi.Work, error) {
	return s.reader.Work(ctx, id, teamID)
}
