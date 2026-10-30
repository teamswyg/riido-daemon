// Package userauth implements standalone CLI OAuth without a Desktop session.
package userauth

import (
	"context"
	"net/http"
	"time"
)

type Endpoints struct{ AuthorizeURL, TokenURL, RevokeURL string }

type Config struct {
	ClientID    string
	Endpoints   Endpoints
	HTTPClient  *http.Client
	Scopes      []string
	Resource    string
	AuthTimeout time.Duration
}

type Credential struct {
	SessionID             string    `json:"session_id"`
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"refresh_token"`
	RefreshRecoverySecret string    `json:"refresh_recovery_secret,omitempty"`
	ExpiresAt             time.Time `json:"expires_at"`
	TokenType             string    `json:"token_type"`
	Scope                 string    `json:"scope,omitempty"`
}

func (Credential) String() string   { return "[redacted OAuth credential]" }
func (Credential) GoString() string { return "[redacted OAuth credential]" }

// CredentialStore must enforce OS user presence on credential access. Lock
// serializes the whole read/refresh/save transaction across CLI processes.
// Ordinary file permissions or an environment variable are not that boundary.
type CredentialStore interface {
	Lock(context.Context) (func(), error)
	Load(context.Context) (Credential, error)
	Save(context.Context, Credential) error
	Delete(context.Context) error
}
