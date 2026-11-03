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
	Binding               *CredentialBinding `json:"binding,omitempty"`
	SessionID             string             `json:"session_id"`
	AccessToken           string             `json:"access_token"`
	RefreshToken          string             `json:"refresh_token"`
	RefreshRecoverySecret string             `json:"refresh_recovery_secret,omitempty"`
	ExpiresAt             time.Time          `json:"expires_at"`
	TokenType             string             `json:"token_type"`
	Scope                 string             `json:"scope,omitempty"`
}

func (Credential) String() string   { return "[redacted OAuth credential]" }
func (Credential) GoString() string { return "[redacted OAuth credential]" }

// CredentialStore must enforce OS user presence on credential access. Lock
// serializes the whole read/refresh/save transaction across CLI processes.
// Ordinary file permissions or an environment variable are not that boundary.
// A store may also implement AccessCache; the protected credential itself
// keeps this contract.
type CredentialStore interface {
	Lock(context.Context) (func(), error)
	Load(context.Context) (Credential, error)
	Save(context.Context, Credential) error
	Delete(context.Context) error
}

// AccessGrant is the short-lived part of a Credential: no refresh token and no
// recovery secret.
type AccessGrant struct {
	Binding     *CredentialBinding `json:"binding,omitempty"`
	SessionID   string             `json:"session_id"`
	AccessToken string             `json:"access_token"`
	ExpiresAt   time.Time          `json:"expires_at"`
	TokenType   string             `json:"token_type"`
	Scope       string             `json:"scope,omitempty"`
}

func (AccessGrant) String() string   { return "[redacted OAuth access grant]" }
func (AccessGrant) GoString() string { return "[redacted OAuth access grant]" }

func (c Credential) AccessGrant() AccessGrant {
	return AccessGrant{Binding: c.Binding, SessionID: c.SessionID, AccessToken: c.AccessToken, ExpiresAt: c.ExpiresAt, TokenType: c.TokenType, Scope: c.Scope}
}

func (g AccessGrant) credential() Credential {
	return Credential{Binding: g.Binding, SessionID: g.SessionID, AccessToken: g.AccessToken, ExpiresAt: g.ExpiresAt, TokenType: g.TokenType, Scope: g.Scope}
}

// AccessCache is an optional CredentialStore extension (RIID-6014). It keeps
// only the current access grant in OS-protected, this-device storage without a
// user-presence prompt, so ordinary reads do not ask for confirmation while the
// access token is valid. Refresh, login and installation-key signing still go
// through the user-presence protected credential. The store keeps the cache in
// step with Save and Delete; Manager reads it under the same Lock.
type AccessCache interface {
	LoadAccess(context.Context) (AccessGrant, error)
	SaveAccess(context.Context, AccessGrant) error
}
