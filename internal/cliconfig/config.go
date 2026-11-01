// Package cliconfig stores nonsecret user CLI preferences, never credentials or grants.
package cliconfig

import (
	"errors"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/internal/userauth"
)

var (
	ErrInvalid     = errors.New("invalid CLI settings")
	ErrUnavailable = errors.New("CLI settings unavailable")
	ErrConflict    = errors.New("CLI settings changed during command")
	idPattern      = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)
)

type Config struct {
	Version       int                `json:"version"`
	ActiveProfile string             `json:"activeProfile"`
	Profiles      map[string]Profile `json:"profiles"`
}

type Profile struct {
	APIOrigin       string `json:"apiOrigin,omitempty"`
	AuthOrigin      string `json:"authOrigin,omitempty"`
	SelectionOrigin string `json:"selectionOrigin,omitempty"`
	Workspace       string `json:"workspace,omitempty"`
	Team            string `json:"team,omitempty"`
	Format          string `json:"format,omitempty"`
	Timeout         string `json:"timeout,omitempty"`
}

func (p Profile) AuthenticationOrigin() string {
	if p.AuthOrigin == "" {
		return userauth.AuthOriginBrowser
	}
	return p.AuthOrigin
}

func Default() Config {
	return Config{Version: 1, ActiveProfile: "production", Profiles: map[string]Profile{}}
}

func ValidProfile(name string) bool {
	return name == "production" || name == "staging" || name == "local"
}
func ValidID(value string) bool { return idPattern.MatchString(value) }
func ValidAuthOrigin(value string) bool {
	return value == userauth.AuthOriginBrowser || value == userauth.AuthOriginDesktop
}

func Origin(raw string, localOnly bool) (string, error) {
	origin := strings.TrimRight(strings.TrimSpace(raw), "/")
	_, err := userauth.NewClient(userauth.Config{ClientID: "riido-cli", Resource: origin, Scopes: []string{"read"}, Endpoints: userauth.Endpoints{AuthorizeURL: origin + "/oauth/authorize", TokenURL: origin + "/oauth/token", RevokeURL: origin + "/oauth/revoke"}})
	if err != nil {
		return "", ErrInvalid
	}
	if localOnly {
		u, _ := url.Parse(origin)
		ip := net.ParseIP(u.Hostname())
		if ip == nil || !ip.IsLoopback() {
			return "", ErrInvalid
		}
	}
	return origin, nil
}

func ValidTimeout(value string) bool {
	d, err := time.ParseDuration(value)
	return err == nil && d > 0 && d <= 10*time.Minute
}

func (c Config) Valid() bool {
	if c.Version != 1 || !ValidProfile(c.ActiveProfile) || c.Profiles == nil || len(c.Profiles) > 3 {
		return false
	}
	for name, p := range c.Profiles {
		if !ValidProfile(name) || p.AuthOrigin != "" && !ValidAuthOrigin(p.AuthOrigin) {
			return false
		}
		if p.APIOrigin != "" {
			o, err := Origin(p.APIOrigin, name == "local")
			if err != nil || o != p.APIOrigin {
				return false
			}
		}
		if p.SelectionOrigin != "" {
			o, err := Origin(p.SelectionOrigin, name == "local")
			if err != nil || o != p.SelectionOrigin {
				return false
			}
		}
		if p.Workspace != "" && !ValidID(p.Workspace) || p.Team != "" && !ValidID(p.Team) {
			return false
		}
		if (p.Workspace != "" || p.Team != "") && p.SelectionOrigin == "" {
			return false
		}
		if p.Format != "" && p.Format != "json" && p.Format != "text" || p.Timeout != "" && !ValidTimeout(p.Timeout) {
			return false
		}
	}
	return true
}
